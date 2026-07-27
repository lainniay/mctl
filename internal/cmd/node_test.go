package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appconfig "github.com/lainniay/mctl/internal/config"
)

func TestNodeUseRejectsAutomaticGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/group":
			_, _ = fmt.Fprint(w, `{"proxies":[{"name":"AutoTest","type":"URLTest","now":"Japan 01","all":["Japan 01","Japan 02"],"alive":true}]}`)
		case "/proxies/Japan 02/delay":
			_, _ = fmt.Fprint(w, `{"delay":83}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MIHOMO_URL", server.URL)
	t.Setenv("MIHOMO_SECRET", "secret")
	cfg := appconfig.Config{
		Version:      2,
		DefaultGroup: "AutoTest",
		Groups: []appconfig.Group{{
			Name:       "AutoTest",
			Type:       "url-test",
			Selectable: true,
			URL:        defaultDelayURL,
			Interval:   300,
		}},
		Rules: []appconfig.Rule{{Match: "MATCH", Policy: "AutoTest"}},
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	nodeUseCmd.SetContext(t.Context())
	t.Cleanup(func() { nodeUseCmd.SetContext(context.Background()) })
	err := nodeUseCmd.RunE(nodeUseCmd, []string{"Japan 02"})
	if err == nil || !strings.Contains(err.Error(), "selects nodes automatically") {
		t.Fatalf("node use error = %v, want automatic group rejection", err)
	}
	var output bytes.Buffer
	nodeTestCmd.SetContext(t.Context())
	nodeTestCmd.SetOut(&output)
	t.Cleanup(func() {
		nodeTestCmd.SetContext(context.Background())
		nodeTestCmd.SetOut(nil)
	})
	if err := nodeTestCmd.RunE(nodeTestCmd, []string{"Japan 02"}); err != nil {
		t.Fatalf("node test failed for automatic group: %v", err)
	}
	if output.String() != "Japan 02\t83ms\n" {
		t.Fatalf("node test output = %q", output.String())
	}
}

func TestNodeUseSelectsMemberOfSelector(t *testing.T) {
	selected := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/group":
			_, _ = fmt.Fprint(w, `{"proxies":[{"name":"NodeSelection","type":"Selector","now":"Japan 01","all":["Japan 01","Japan 02"],"alive":true}]}`)
		case r.Method == http.MethodPut && r.URL.Path == "/proxies/NodeSelection":
			selected = "Japan 02"
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MIHOMO_URL", server.URL)
	t.Setenv("MIHOMO_SECRET", "secret")
	cfg := appconfig.Config{
		Version:      2,
		DefaultGroup: "NodeSelection",
		Groups:       []appconfig.Group{{Name: "NodeSelection", Type: "select", Selectable: true}},
		Rules:        []appconfig.Rule{{Match: "MATCH", Policy: "NodeSelection"}},
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	nodeUseCmd.SetContext(t.Context())
	nodeUseCmd.SetOut(&output)
	t.Cleanup(func() {
		nodeUseCmd.SetContext(context.Background())
		nodeUseCmd.SetOut(nil)
	})
	if err := nodeUseCmd.RunE(nodeUseCmd, []string{"Japan 02"}); err != nil {
		t.Fatalf("node use failed: %v", err)
	}
	if selected != "Japan 02" || output.String() != "current node: Japan 02\n" {
		t.Fatalf("selected = %q, output = %q", selected, output.String())
	}
}
