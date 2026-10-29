package sub

import (
	"reflect"
	"strings"
	"testing"
)

func TestVLESSRegionNamePreservesProtocol(t *testing.T) {
	proxy, err := ParseURL("vless://uuid@example.com:443?type=ws&security=tls#JP-VLESS")
	if err != nil {
		t.Fatal(err)
	}
	got := Clean([]Proxy{proxy})
	if len(got) != 1 || got[0].Name != "Japan-01-VLESS" {
		t.Fatalf("Clean() = %+v, want Japan-01-VLESS", got)
	}
}

func TestParseVLESSRejectsInvalidEndpoints(t *testing.T) {
	for _, raw := range []string{
		"vless://example.com:443#JP",
		"vless://uuid@:443#JP",
		"vless://uuid@example.com#JP",
		"vless://uuid@example.com:0#JP",
		"vless://uuid@example.com:65536#JP",
		"vless://uuid:password@example.com:443#JP",
		"vless://uuid@2001:db8::1:443#JP",
		"vless://uuid@example.com:443/path#JP",
		"vless://uuid@example.com:443?type=grpc#JP",
		"vless://uuid@example.com:443?security=reality&pbk=public-key#JP",
		"vless://uuid@example.com:443?type=tcp&path=%2Fws#JP",
		"vless://uuid@example.com:443?insecure=maybe#JP",
		"vless://uuid@example.com:443?insecure=#JP",
		"vless://uuid@example.com:443?security=tls&fp=#JP",
		"vless://uuid@example.com:443?security=tls&fp=invalid#JP",
		"vless://uuid@example.com:443?security=tls&fp=Chrome#JP",
		"vless://uuid@example.com:443?security=none&fp=chrome#JP",
		"vless://uuid@example.com:443?type=ws&type=tcp#JP",
		"vless://uuid@example.com:443?security=reality&sni=www.amazon.co.jp&pbk=short#JP",
		"vless://uuid@example.com:443?security=reality&sni=www.amazon.co.jp&pbk=rS8km7nWSrTotcgYGP_wrqWAkhzz4Q6E6BcK4RM7qE4&sid=zzzz#JP",
		"vless://uuid@example.com:443?pbk=rS8km7nWSrTotcgYGP_wrqWAkhzz4Q6E6BcK4RM7qE4#JP",
		"vless://uuid@example.com:443?sid=0b69a2cc511632a8#JP",
		"vless://uuid@example.com:443?flow=#JP",
		"vless://uuid@example.com:443?flow=xtls-rprx-direct#JP",
		"vless://uuid@example.com:443?type=ws&flow=xtls-rprx-vision#JP",
	} {
		if _, err := ParseURL(raw); err == nil {
			t.Errorf("accepted invalid VLESS URI %q", raw)
		}
	}

	_, err := ParseURL("vless://uuid@example.com:443?%1B%5D52%3Bc%3Bpayload=1#JP")
	if err == nil || strings.Contains(err.Error(), "\x1b") {
		t.Fatalf("unsafe VLESS error: %q", err)
	}
}

func TestParseVLESSClientFingerprints(t *testing.T) {
	for _, fingerprint := range []string{
		"chrome", "firefox", "safari", "ios", "android", "edge", "360", "qq", "random",
		"chrome120", "firefox120", "safari16", "chrome_psk", "chrome_psk_shuffle",
		"chrome_padding_psk_shuffle", "chrome_pq", "chrome_pq_psk", "randomized",
	} {
		proxy, err := ParseURL("vless://uuid@example.com:443?security=tls&fp=" + fingerprint)
		if err != nil {
			t.Errorf("fp=%q: %v", fingerprint, err)
			continue
		}
		if proxy.ClientFingerprint != fingerprint {
			t.Errorf("fp=%q parsed as %q", fingerprint, proxy.ClientFingerprint)
		}
	}

	proxy, err := ParseURL("vless://uuid@example.com:443?fp=none")
	if err != nil {
		t.Fatal(err)
	}
	if proxy.ClientFingerprint != "" {
		t.Errorf("fp=none parsed as %q, want empty", proxy.ClientFingerprint)
	}
}

func TestParseVLESSPreservesRealityAndFlowOptions(t *testing.T) {
	raw := "vless://9873aa73-c4c7-4996-a39e-a0d56a2a6ddd@100.70.224.56:24444?encryption=none&flow=xtls-rprx-vision&security=reality&sni=www.amazon.co.jp&fp=chrome&pbk=rS8km7nWSrTotcgYGP_wrqWAkhzz4Q6E6BcK4RM7qE4&sid=0b69a2cc511632a8&type=tcp&headerType=none#JP-VLESS-REALITY"

	got, err := ParseURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := Proxy{
		Name:              "JP-VLESS-REALITY",
		Type:              "vless",
		Server:            "100.70.224.56",
		Port:              24444,
		UUID:              "9873aa73-c4c7-4996-a39e-a0d56a2a6ddd",
		SNI:               "www.amazon.co.jp",
		Network:           "tcp",
		Security:          "reality",
		ClientFingerprint: "chrome",
		Flow:              "xtls-rprx-vision",
		RealityPublicKey:  "rS8km7nWSrTotcgYGP_wrqWAkhzz4Q6E6BcK4RM7qE4",
		RealityShortID:    "0b69a2cc511632a8",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseURL() = %+v, want %+v", got, want)
	}
}

func TestVLESSDeduplicationIncludesFlowAndReality(t *testing.T) {
	base := "vless://uuid@example.com:443?security=reality&sni=www.amazon.co.jp&pbk=rS8km7nWSrTotcgYGP_wrqWAkhzz4Q6E6BcK4RM7qE4"

	flowed, err := ParseURL(base + "&flow=xtls-rprx-vision#A")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := ParseURL(base + "#B")
	if err != nil {
		t.Fatal(err)
	}
	if flowed.Equal(plain) {
		t.Error("flow is not part of VLESS connection identity")
	}

	withSID, err := ParseURL(base + "&sid=0b69a2cc511632a8#C")
	if err != nil {
		t.Fatal(err)
	}
	if withSID.Equal(plain) {
		t.Error("REALITY short ID is not part of VLESS connection identity")
	}
}

func TestParseVLESSPreservesWebSocketOptions(t *testing.T) {
	raw := "vless://uuid@example.com:443?type=ws&security=tls&encryption=none&sni=edge.example.com&fp=chrome&path=%2Fsocket%3Fed%3D2048&host=cdn.example.com&insecure=1#Cloudflare"

	got, err := ParseURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := Proxy{
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
		SkipCertVerify:    true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseURL() = %+v, want %+v", got, want)
	}
}
