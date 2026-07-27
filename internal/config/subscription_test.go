package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSubscriptionsRoundTripAndOperations(t *testing.T) {
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	subs := Subs{}
	if err := subs.AddSub("primary", "https://example.com/sub?token=secret"); err != nil {
		t.Fatal(err)
	}
	if err := subs.AddSub("primary", "https://example.com/other"); !errors.Is(err, ErrSubExists) {
		t.Fatalf("duplicate error = %v", err)
	}
	if err := subs.AddSub("secondary", "https://example.com/sub?token=secret"); !errors.Is(err, ErrSubExists) {
		t.Fatalf("duplicate URL error = %v", err)
	}
	if !subs.SetSubEnabled("primary", false) {
		t.Fatal("SetSubEnabled() = false")
	}
	if err := subs.Save(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(stateHome, "mihomo", "subscriptions.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = %o, want 600", info.Mode().Perm())
	}
	got, err := LoadSubs()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].Enabled || got.Items[0].Name != "primary" {
		t.Fatalf("LoadSubs() = %+v", got)
	}
	if !got.RemoveSub("primary") || len(got.Items) != 0 {
		t.Fatalf("RemoveSub() = %+v", got)
	}
}
