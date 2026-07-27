package cmd

import (
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
