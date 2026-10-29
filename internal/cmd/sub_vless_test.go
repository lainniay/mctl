package cmd

import (
	"os"
	"testing"

	"github.com/lainniay/mctl/internal/config"
	"go.yaml.in/yaml/v4"
)

func TestSubAddDirectVLESSAndUpdate(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	raw := "vless://uuid@example.com:443?type=ws&security=tls&sni=edge.example.com&fp=chrome&path=%2Fws&host=cdn.example.com#JP-VLESS"
	t.Setenv("MCTL_SUB_URL", raw)

	if err := subAddCmd.RunE(subAddCmd, []string{"personal-vless"}); err != nil {
		t.Fatal(err)
	}
	subs, err := config.LoadSubs()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs.Items) != 1 || subs.Items[0].URL != raw || !subs.Items[0].Enabled {
		t.Fatal("direct VLESS source not persisted")
	}

	count, path, err := runSubUpdate()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d, want 1", count)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Proxies []struct {
			Name, Type, Server, UUID, Network, ServerName string
			ClientFingerprint                             string `yaml:"client-fingerprint"`
			Port                                          int
			TLS                                           bool
		} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Proxies) != 1 {
		t.Fatalf("proxies=%d, want 1", len(got.Proxies))
	}
	proxy := got.Proxies[0]
	if proxy.Name != "Japan-01-VLESS" || proxy.Type != "vless" || proxy.Server != "example.com" || proxy.Port != 443 || proxy.UUID != "uuid" || proxy.Network != "ws" || proxy.ServerName != "edge.example.com" || proxy.ClientFingerprint != "chrome" || !proxy.TLS {
		t.Fatalf("VLESS config not preserved: %+v", proxy)
	}
}
