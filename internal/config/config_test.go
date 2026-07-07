package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_returnsEmptyConfig_whenFileMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Subs) != 0 {
		t.Fatalf("Load() subs = %v, want empty", cfg.Subs)
	}
}

func TestSave_createsConfigDirectoryAndWritesIndentedJSON(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	cfg := Config{Subs: []Sub{{Name: "naiyun", URL: "https://example.com/sub", Enabled: true}}}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(configHome, "mihomo", "mctl.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "\n  \"subs\":") {
		t.Fatalf("Save() wrote %q, want indented JSON", data)
	}
}

func TestAddSub_rejectsDuplicateNameOrURL(t *testing.T) {
	cfg := Config{}

	if err := cfg.AddSub("naiyun", "https://example.com/sub"); err != nil {
		t.Fatalf("AddSub() error = %v", err)
	}
	if err := cfg.AddSub("naiyun", "https://example.com/other"); !errors.Is(err, ErrSubExists) {
		t.Fatalf("AddSub() duplicate name error = %v, want ErrSubExists", err)
	}
	if err := cfg.AddSub("other", "https://example.com/sub"); !errors.Is(err, ErrSubExists) {
		t.Fatalf("AddSub() duplicate URL error = %v, want ErrSubExists", err)
	}
}

func TestRemoveSub_removesByNameOrURL(t *testing.T) {
	cfg := Config{Subs: []Sub{
		{Name: "naiyun", URL: "https://example.com/naiyun", Enabled: true},
		{Name: "other", URL: "https://example.com/other", Enabled: true},
	}}

	if !cfg.RemoveSub("naiyun") {
		t.Fatal("RemoveSub() by name = false, want true")
	}
	if !cfg.RemoveSub("https://example.com/other") {
		t.Fatal("RemoveSub() by URL = false, want true")
	}
	if cfg.RemoveSub("missing") {
		t.Fatal("RemoveSub() missing = true, want false")
	}
}
