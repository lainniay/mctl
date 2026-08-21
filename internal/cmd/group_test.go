package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	appconfig "github.com/lainniay/mctl/internal/config"
	"github.com/lainniay/mctl/internal/mihomo"
)

func TestSelectableGroupsUsesMctlMetadataNotRuntimeType(t *testing.T) {
	cfg := appconfig.Config{Groups: []appconfig.Group{
		{Name: "AutoTest", Type: "url-test", Selectable: true},
		{Name: "OpenAI", Type: "select", Selectable: false},
	}}
	runtime := []mihomo.Group{
		{Name: "OpenAI", Type: "Selector"},
		{Name: "AutoTest", Type: "URLTest"},
		{Name: "GLOBAL", Type: "Selector"},
	}

	groups, err := selectableGroups(cfg, runtime)
	if err != nil {
		t.Fatalf("selectableGroups() error = %v", err)
	}
	if len(groups) != 1 || groups[0].Name != "AutoTest" {
		t.Fatalf("selectableGroups() = %+v", groups)
	}
}

func TestGroupChangeUsesSingleGroupEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.EscapedPath() != "/proxies/NodeSelection" {
			t.Errorf("request = %s %s", r.Method, r.URL.EscapedPath())
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprint(w, `{"name":"NodeSelection","type":"Selector","now":"Japan 01","all":["Japan 01"]}`)
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
	groupChangeCmd.SetContext(t.Context())
	groupChangeCmd.SetOut(&output)
	t.Cleanup(func() {
		groupChangeCmd.SetContext(context.Background())
		groupChangeCmd.SetOut(nil)
	})
	if err := groupChangeCmd.RunE(groupChangeCmd, []string{"NodeSelection"}); err != nil {
		t.Fatalf("group change failed: %v", err)
	}
	state, err := appconfig.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if state.CurrentGroup != "NodeSelection" || output.String() != "current group: NodeSelection\n" {
		t.Fatalf("state = %+v, output = %q", state, output.String())
	}
}
