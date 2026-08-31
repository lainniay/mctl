package sub

import (
	"reflect"
	"testing"
)

func TestParseVLESSPreservesWebSocketOptions(t *testing.T) {
	raw := "vless://uuid@example.com:443?type=ws&security=tls&sni=edge.example.com&path=%2Fsocket%3Fed%3D2048&host=cdn.example.com&insecure=1#Cloudflare"

	got, err := ParseURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := Proxy{
		Name:           "Cloudflare",
		Type:           "vless",
		Server:         "example.com",
		Port:           443,
		UUID:           "uuid",
		SNI:            "edge.example.com",
		Network:        "ws",
		Security:       "tls",
		WSPath:         "/socket?ed=2048",
		WSHost:         "cdn.example.com",
		SkipCertVerify: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseURL() = %+v, want %+v", got, want)
	}
}
