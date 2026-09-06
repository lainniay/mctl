package sub

import (
	"encoding/base64"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func ss2022URI(cipher string, keyBytes int) string {
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", keyBytes)))
	return "ss://" + url.UserPassword(cipher, key).String() + "@example.com:24443#JP-SS2022"
}

func TestSSRegionNamePreservesProtocol(t *testing.T) {
	proxy, err := ParseURL(ss2022URI("2022-blake3-aes-256-gcm", 32))
	if err != nil {
		t.Fatal(err)
	}
	got := Clean([]Proxy{proxy})
	if len(got) != 1 || got[0].Name != "Japan-01-SS" {
		t.Fatalf("Clean() = %+v, want Japan-01-SS", got)
	}
}

func TestParseSS2022(t *testing.T) {
	for cipher, size := range map[string]int{"2022-blake3-aes-128-gcm": 16, "2022-blake3-aes-256-gcm": 32, "2022-blake3-chacha20-poly1305": 32} {
		t.Run(cipher, func(t *testing.T) {
			got, err := ParseURL(ss2022URI(cipher, size))
			if err != nil {
				t.Fatal(err)
			}
			want := Proxy{Name: "JP-SS2022", Type: "ss", Server: "example.com", Port: 24443, Password: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", size))), SS: ShadowsocksOptions{Cipher: cipher, UDP: true}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %+v, want %+v", got, want)
			}
		})
	}
	// AES EIH identity keys are part of the password, not URI separators.
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", 32)))
	raw := "ss://" + url.UserPassword("2022-blake3-aes-256-gcm", key+":"+key).String() + "@example.com:443"
	p, err := ParseURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.Password != key+":"+key {
		t.Fatal("key chain was not preserved")
	}
}

func TestParseSSFormats(t *testing.T) {
	credentials := "aes-256-gcm:p@ss:/+?#%word"
	want := Proxy{Name: "JP node", Type: "ss", Server: "2001:db8::1", Port: 443, Password: "p@ss:/+?#%word", SS: ShadowsocksOptions{Cipher: "aes-256-gcm", UDP: true}}
	links := []string{"ss://" + url.UserPassword("aes-256-gcm", want.Password).String() + "@[2001:db8::1]:443/#JP%20node"}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		links = append(links, "ss://"+enc.EncodeToString([]byte(credentials))+"@[2001:db8::1]:443#JP%20node")
		links = append(links, "ss://"+enc.EncodeToString([]byte(credentials+"@[2001:db8::1]:443"))+"#JP%20node")
	}
	for _, raw := range links {
		p, err := ParseURL(raw)
		if err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		if p != want {
			t.Fatalf("got %+v, want %+v", p, want)
		}
	}
	body := strings.Join([]string{links[0], links[1]}, "\n")
	for _, input := range []string{body, base64.StdEncoding.EncodeToString([]byte(body))} {
		got, err := Parse(input)
		if err != nil || len(got) != 2 {
			t.Fatalf("Parse: %v, %v", got, err)
		}
		if len(Clean(got)) != 1 {
			t.Fatal("equivalent SS encodings did not dedupe")
		}
	}
}

func TestParseSSRejectsInvalidWithoutLeakingCredentials(t *testing.T) {
	base := "ss://aes-256-gcm:private-password@example.com:443"
	cases := []string{
		"ss://", "ss://bad-base64@example.com:443", "ss://" + base64.StdEncoding.EncodeToString([]byte("missing separator")),
		"ss://aes-256-gcm:@example.com:443", "ss://:private-password@example.com:443",
		"ss://aes-256-gcm:private-password@:443", "ss://aes-256-gcm:private-password@example.com",
		"ss://aes-256-gcm:private-password@example.com:0", "ss://aes-256-gcm:private-password@example.com:65536",
		"ss://aes-256-gcm:private-password@example.com:abc", "ss://aes-256-gcm:private-password@2001:db8::1:443",
		base + "/path", base + "?unknown=private-password", base + "?udp=maybe", base + "?udp=true&udp=false", base + "?udp=%XX",
		base + "?plugin=unknown", base + "?plugin=obfs-local", base + "?plugin=" + url.QueryEscape("obfs-local;obfs=wrong"),
		base + "?plugin=" + url.QueryEscape("obfs-local;obfs=tls;unknown=x"),
		base + "?plugin=" + url.QueryEscape("v2ray-plugin;mode=quic"),
		base + "?plugin=" + url.QueryEscape("v2ray-plugin;tls=maybe"),
		base + "?plugin=" + url.QueryEscape("v2ray-plugin;tls;tls"),
		base + "?plugin=" + url.QueryEscape(`v2ray-plugin;host=escaped\;host`),
		base + "?udp-over-tcp-version=2", base + "?udp-over-tcp=true&udp-over-tcp-version=3",
		"ss://2022-blake3-aes-256-gcm:private-password@example.com:443",
		ss2022URI("2022-blake3-aes-128-gcm", 32), ss2022URI("2022-unknown", 32),
		"ss://" + base64.RawURLEncoding.EncodeToString([]byte("2022-blake3-aes-256-gcm:"+base64.StdEncoding.EncodeToString(make([]byte, 32)))) + "@example.com:443",
		"ss://aes-256-gcm:private-password%XX@example.com:443",
	}
	for _, raw := range cases {
		if _, err := ParseURL(raw); err == nil {
			t.Errorf("accepted %s", raw)
		} else if strings.Contains(err.Error(), "private-password") {
			t.Errorf("credentials leaked: %v", err)
		}
	}
}

func TestSSPluginRenderingAndIdentity(t *testing.T) {
	base := "ss://aes-256-gcm:password@example.com:443"
	tests := []struct {
		query, plugin string
		opts          map[string]any
		udp           bool
		uot           bool
		version       int
	}{
		{"", "", nil, true, false, 0},
		{"plugin=" + url.QueryEscape("obfs-local;obfs=tls"), "obfs", map[string]any{"mode": "tls", "host": "cloudfront.net"}, true, false, 0},
		{"plugin=v2ray-plugin", "v2ray-plugin", map[string]any{"mode": "websocket", "host": "cloudfront.com", "path": "/", "tls": false, "mux": true, "skip-cert-verify": false}, true, false, 0},
		{"udp=false", "", nil, false, false, 0},
		{"udp-over-tcp=true", "", nil, true, true, 2},
		{"udp-over-tcp=true&udp-over-tcp-version=1", "", nil, true, true, 1},
		{"plugin=" + url.QueryEscape("obfs-local;obfs=tls;obfs-host=cdn.example.com"), "obfs", map[string]any{"mode": "tls", "host": "cdn.example.com"}, true, false, 0},
		{"plugin=" + url.QueryEscape("v2ray-plugin;tls;host=cdn.example.com;path=/ws?ed=2048;mux=false;skip-cert-verify"), "v2ray-plugin", map[string]any{"mode": "websocket", "tls": true, "host": "cdn.example.com", "path": "/ws?ed=2048", "mux": false, "skip-cert-verify": true}, true, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			p, err := ParseURL(base + "?" + tt.query)
			if err != nil {
				t.Fatal(err)
			}
			data, err := RenderProvider([]Proxy{p})
			if err != nil {
				t.Fatal(err)
			}
			var file providerFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				t.Fatal(err)
			}
			got := file.Proxies[0]
			if got.Cipher != "aes-256-gcm" || got.Password != "password" || got.UDP != tt.udp || got.Plugin != tt.plugin || got.UDPOverTCP != tt.uot || got.UDPOverTCPVersion != tt.version || !reflect.DeepEqual(got.PluginOpts, tt.opts) {
				t.Fatalf("unexpected provider: %s", data)
			}
		})
	}
	original, _ := ParseURL(base)
	if original.Name != "example.com:443" {
		t.Fatalf("untagged SS name = %q", original.Name)
	}
	left, err := ParseURL(base + "?plugin=" + url.QueryEscape("obfs-local;obfs=tls;obfs-host=cdn.example.com"))
	if err != nil {
		t.Fatal(err)
	}
	right, err := ParseURL(base + "?plugin=" + url.QueryEscape("obfs;obfs-host=cdn.example.com;obfs=tls") + "#Other")
	if err != nil {
		t.Fatal(err)
	}
	if !left.Equal(right) {
		t.Fatal("plugin aliases and option order must not affect identity")
	}
	variations := []func(*Proxy){
		func(p *Proxy) { p.SS.Cipher = "aes-128-gcm" }, func(p *Proxy) { p.Password = "other" }, func(p *Proxy) { p.SS.UDP = false },
		func(p *Proxy) { p.SS.Plugin = "obfs" }, func(p *Proxy) { p.SS.PluginOptions.Mode = "tls" }, func(p *Proxy) { p.SS.PluginOptions.Host = "host" },
		func(p *Proxy) { p.SS.PluginOptions.Path = "/ws" }, func(p *Proxy) { p.SS.PluginOptions.TLS = true }, func(p *Proxy) { p.SS.PluginOptions.Mux = true }, func(p *Proxy) { p.SS.PluginOptions.SkipCertVerify = true },
		func(p *Proxy) { p.SS.UDPOverTCP = true }, func(p *Proxy) { p.SS.UDPOverTCPVersion = 2 },
	}
	for _, change := range variations {
		other := original
		change(&other)
		if original.Equal(other) || len(Clean([]Proxy{original, other})) != 2 {
			t.Fatal("SS connection setting omitted from identity")
		}
	}
	renamed := original
	renamed.Name = "Other"
	if !original.Equal(renamed) {
		t.Fatal("name must not affect identity")
	}
}
