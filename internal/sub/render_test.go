package sub

import (
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestRenderProviderIncludesVLESSWebSocketOptions(t *testing.T) {
	data, err := RenderProvider([]Proxy{{
		Name:     "Cloudflare",
		Type:     "vless",
		Server:   "example.com",
		Port:     443,
		UUID:     "uuid",
		SNI:      "edge.example.com",
		Network:  "ws",
		Security: "tls",
		WSPath:   "/socket?ed=2048",
		WSHost:   "cdn.example.com",
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
