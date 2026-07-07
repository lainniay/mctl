package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lainniay/mctl/internal/config"
)

func Test_runSubUpdate_writesProvider_whenEnabledSubscriptionsExist(t *testing.T) {
	// Given
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "anytls://secret@hk.example.com:443?sni=hk.example.com&insecure=1#HK%2001\n"); err != nil {
			return
		}
	}))
	defer server.Close()

	cfg := config.Config{Subs: []config.Sub{
		{Name: "enabled", URL: server.URL, Enabled: true},
		{Name: "disabled", URL: "http://127.0.0.1:1/nope", Enabled: false},
	}}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	// When
	count, path, err := runSubUpdate()

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 proxy, got %d", count)
	}
	wantPath := filepath.Join(configHome, "mihomo", "providers", "nodes.yaml")
	if path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, path)
	}
	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"proxies:", "name: Hong Kong 01", "type: anytls", "skip-cert-verify: true"} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected provider to contain %q, got:\n%s", want, content)
		}
	}
}
