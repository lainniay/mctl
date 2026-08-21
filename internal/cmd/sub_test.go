package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lainniay/mctl/internal/config"
)

func TestSubListDoesNotPrintURLs(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	subs := config.Subs{Items: []config.Sub{{
		Name: "primary", URL: "https://example.com/sub?token=secret", Enabled: true,
	}}}
	if err := subs.Save(); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	subListCmd.SetOut(&output)
	t.Cleanup(func() { subListCmd.SetOut(nil) })
	if err := subListCmd.RunE(subListCmd, nil); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "primary\ttrue\n" {
		t.Fatalf("sub list output = %q", got)
	}
}

func TestSubAddStoresTemporaryEnvironmentURL(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MCTL_SUB_URL", "https://example.com/sub?token=secret")

	if err := subAddCmd.RunE(subAddCmd, []string{"primary"}); err != nil {
		t.Fatal(err)
	}
	subs, err := config.LoadSubs()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs.Items) != 1 || subs.Items[0].Name != "primary" || subs.Items[0].URL == "" {
		t.Fatalf("LoadSubs() = %+v", subs)
	}
}

func Test_runSubUpdate_writesProvider_whenEnabledSubscriptionsExist(t *testing.T) {
	// Given
	stateHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", stateHome)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "anytls://secret@hk.example.com:443?sni=hk.example.com&insecure=1#HK%2001\n"); err != nil {
			return
		}
	}))
	defer server.Close()

	subs := config.Subs{Items: []config.Sub{
		{Name: "enabled", URL: server.URL, Enabled: true},
		{Name: "disabled", URL: "http://127.0.0.1:1/nope", Enabled: false},
	}}
	if err := subs.Save(); err != nil {
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
	wantPath := filepath.Join(stateHome, "mihomo", "providers", "nodes.yaml")
	if path != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, path)
	}
	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"proxies:", "name: HongKong-01", "type: anytls", "skip-cert-verify: true"} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected provider to contain %q, got:\n%s", want, content)
		}
	}
}
