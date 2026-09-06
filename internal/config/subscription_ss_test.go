package config

import (
	"errors"
	"testing"
)

func TestAddSSSourceValidationAndDuplicates(t *testing.T) {
	raw := "ss://aes-256-gcm:test-password@example.com:443#JP"
	var subs Subs
	if err := subs.AddSub("personal", raw); err != nil {
		t.Fatal(err)
	}
	if err := subs.AddSub("duplicate", raw); !errors.Is(err, ErrSubExists) {
		t.Fatalf("duplicate URI: %v", err)
	}
	if err := subs.AddSub("personal", "https://example.com/sub"); !errors.Is(err, ErrSubExists) {
		t.Fatalf("duplicate name: %v", err)
	}
	for _, raw := range []string{"file:///tmp/nodes", "ss://aes-256-gcm:password@example.com:0", "ssr://test"} {
		if err := subs.AddSub("invalid", raw); err == nil {
			t.Fatal("accepted invalid source")
		}
	}
	if len(subs.Items) != 1 {
		t.Fatal("failed adds changed subscriptions")
	}
}
