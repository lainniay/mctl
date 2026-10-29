package sub

import (
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestRenderProviderIncludesVLESSRealityAndFlow(t *testing.T) {
	data, err := RenderProvider([]Proxy{{
		Name:              "JP-VLESS-REALITY",
		Type:              "vless",
		Server:            "100.70.224.56",
		Port:              24444,
		UUID:              "uuid",
		SNI:               "www.amazon.co.jp",
		Network:           "tcp",
		Security:          "reality",
		ClientFingerprint: "chrome",
		Flow:              "xtls-rprx-vision",
		RealityPublicKey:  "rS8km7nWSrTotcgYGP_wrqWAkhzz4Q6E6BcK4RM7qE4",
		RealityShortID:    "0b69a2cc511632a8",
	}})
	if err != nil {
		t.Fatal(err)
	}

	var got providerFile
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	proxy := got.Proxies[0]
	if !proxy.TLS {
		t.Error("REALITY proxy must render tls: true")
	}
	if proxy.Flow != "xtls-rprx-vision" {
		t.Errorf("flow = %q", proxy.Flow)
	}
	if proxy.RealityOpts == nil {
		t.Fatal("rendered VLESS proxy is missing reality-opts")
	}
	if proxy.RealityOpts.PublicKey != "rS8km7nWSrTotcgYGP_wrqWAkhzz4Q6E6BcK4RM7qE4" {
		t.Errorf("reality-opts.public-key = %q", proxy.RealityOpts.PublicKey)
	}
	if proxy.RealityOpts.ShortID != "0b69a2cc511632a8" {
		t.Errorf("reality-opts.short-id = %q", proxy.RealityOpts.ShortID)
	}
}

func TestRenderProviderIncludesVLESSWebSocketOptions(t *testing.T) {
	data, err := RenderProvider([]Proxy{{
		Name:              "Cloudflare",
		Type:              "vless",
		Server:            "example.com",
		Port:              443,
		UUID:              "uuid",
		SNI:               "edge.example.com",
		Network:           "ws",
		Security:          "tls",
		ClientFingerprint: "chrome",
		WSPath:            "/socket?ed=2048",
		WSHost:            "cdn.example.com",
	}})
	if err != nil {
		t.Fatal(err)
	}

	var got providerFile
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Proxies) != 1 {
		t.Fatalf("rendered proxies = %d, want 1", len(got.Proxies))
	}
	proxy := got.Proxies[0]
	if proxy.ClientFingerprint != "chrome" {
		t.Errorf("client-fingerprint = %q", proxy.ClientFingerprint)
	}
	if proxy.WSOpts == nil {
		t.Fatal("rendered VLESS proxy is missing ws-opts")
	}
	if proxy.WSOpts.Path != "/socket?ed=2048" {
		t.Errorf("ws-opts.path = %q", proxy.WSOpts.Path)
	}
	if proxy.WSOpts.Headers["Host"] != "cdn.example.com" {
		t.Errorf("ws-opts.headers.Host = %q", proxy.WSOpts.Headers["Host"])
	}
}
