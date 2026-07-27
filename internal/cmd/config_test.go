package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/lainniay/mctl/internal/config"
)

func TestRenderConfig_usesXdgConfigAndState(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MIHOMO_URL", "127.0.0.1:9090")
	t.Setenv("MIHOMO_SECRET", "test-secret")
	cfg := appconfig.Config{
		Version:      2,
		DefaultGroup: "NodeSelection",
		Groups:       []appconfig.Group{{Name: "NodeSelection", Type: "select", Selectable: true}},
		Rules:        []appconfig.Rule{{Match: "MATCH", Policy: "NodeSelection"}},
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	basePath, err := appconfig.BasePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(basePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(basePath, []byte("port: 7890\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	data, home, err := renderConfig()
	if err != nil {
		t.Fatalf("renderConfig() error = %v", err)
	}
	if !strings.Contains(string(data), "MctlNodes:") {
		t.Fatalf("renderConfig() output:\n%s", data)
	}
	if !strings.HasSuffix(home, "/mihomo") {
		t.Fatalf("renderConfig() home = %q", home)
	}
}

func TestRunConfigApply_installsBacksUpAndReloads(t *testing.T) {
	target, backup, old, requests := setupConfigApply(t, http.StatusNoContent)

	path, err := runConfigApply(t.Context())
	if err != nil {
		t.Fatalf("runConfigApply() error = %v", err)
	}
	if path != target || *requests != 1 {
		t.Fatalf("runConfigApply() = %q, requests = %d", path, *requests)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "MctlNodes:") {
		t.Fatalf("active config was not generated:\n%s", data)
	}
	data, err = os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != old {
		t.Fatalf("backup = %q", data)
	}
}

func TestRunConfigApply_restoresConfigWhenReloadFails(t *testing.T) {
	target, _, old, requests := setupConfigApply(t, http.StatusInternalServerError)

	_, err := runConfigApply(t.Context())
	if err == nil || !strings.Contains(err.Error(), "reload config") {
		t.Fatalf("runConfigApply() error = %v", err)
	}
	data, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != old || *requests != 2 {
		t.Fatalf("rollback config = %q, requests = %d", data, *requests)
	}
}

func TestRunConfigApply_removesNewConfigWhenFirstReloadFails(t *testing.T) {
	target, _, _, requests := setupConfigApply(t, http.StatusInternalServerError)
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}

	_, err := runConfigApply(t.Context())
	if err == nil || !strings.Contains(err.Error(), "reload config") {
		t.Fatalf("runConfigApply() error = %v", err)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("failed first apply left config at %s: %v", target, statErr)
	}
	if *requests != 1 {
		t.Fatalf("reload requests = %d, want 1", *requests)
	}
}

func setupConfigApply(t *testing.T, status int) (target, backup, old string, requests *int) {
	t.Helper()
	configHome := t.TempDir()
	stateHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_STATE_HOME", stateHome)
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	t.Setenv("MIHOMO_URL", server.URL)
	t.Setenv("MIHOMO_SECRET", "test-secret")

	cfg := appconfig.Config{
		Version:      2,
		DefaultGroup: "NodeSelection",
		Groups:       []appconfig.Group{{Name: "NodeSelection", Type: "select", Selectable: true}},
		Rules:        []appconfig.Rule{{Match: "MATCH", Policy: "NodeSelection"}},
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	basePath, err := appconfig.BasePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(basePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(basePath, []byte("mode: rule\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir, err := appconfig.Dir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "providers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "providers", "nodes.yaml"), []byte("proxies: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target = filepath.Join(dir, "config.yaml")
	old = "mode: rule\n" + fmt.Sprintf("external-controller: %q\nsecret: test-secret\n", server.URL)
	if err := os.WriteFile(target, []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}
	backup, err = appconfig.BackupPath()
	if err != nil {
		t.Fatal(err)
	}

	bin := t.TempDir()
	mihomo := filepath.Join(bin, "mihomo")
	if err := os.WriteFile(mihomo, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	return target, backup, old, &count
}

func TestRunConfigValidate_requiresGeneratedProvider(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MIHOMO_URL", "127.0.0.1:9090")
	t.Setenv("MIHOMO_SECRET", "test-secret")
	cfg := appconfig.Config{
		Version:      2,
		DefaultGroup: "NodeSelection",
		Groups:       []appconfig.Group{{Name: "NodeSelection", Type: "select", Selectable: true}},
		Rules:        []appconfig.Rule{{Match: "MATCH", Policy: "NodeSelection"}},
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	basePath, err := appconfig.BasePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(basePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(basePath, []byte("mode: rule\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err = runConfigValidate(t.Context())
	if err == nil || !strings.Contains(err.Error(), "run mctl sub update") {
		t.Fatalf("runConfigValidate() error = %v, want provider guidance", err)
	}
}
