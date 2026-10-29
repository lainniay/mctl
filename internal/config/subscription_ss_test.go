package config

import (
	"errors"
	"testing"
)

func TestAddDirectSourceValidationAndDuplicates(t *testing.T) {
	ssRaw := "ss://aes-256-gcm:test-password@example.com:443#JP"
	var subs Subs
	if err := subs.AddSub("personal", ssRaw); err != nil {
		t.Fatal(err)
	}
	if err := subs.AddSub("duplicate", ssRaw); !errors.Is(err, ErrSubExists) {
		t.Fatalf("duplicate URI: %v", err)
	}
	if err := subs.AddSub("personal", "https://example.com/sub"); !errors.Is(err, ErrSubExists) {
		t.Fatalf("duplicate name: %v", err)
	}
	vlessRaw := "vless://uuid@example.com:443?type=ws&security=tls#JP-VLESS"
	if err := subs.AddSub("vless", vlessRaw); err != nil {
		t.Fatalf("add VLESS source: %v", err)
	}
	for _, raw := range []string{"file:///tmp/nodes", "ss://aes-256-gcm:password@example.com:0", "vless://uuid@example.com:not-a-port", "ssr://test"} {
		if err := subs.AddSub("invalid", raw); err == nil {
			t.Fatal("accepted invalid source")
		}
	}
	if len(subs.Items) != 2 {
		t.Fatal("failed adds changed subscriptions")
	}
}
