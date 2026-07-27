package config

import "testing"

func TestValidateRejectsLegacyAndMissingVersions(t *testing.T) {
	for _, version := range []int{0, 1, 3} {
		if err := (Config{Version: version}).Validate(); err == nil {
			t.Fatalf("Validate() version %d error = nil", version)
		}
	}
}

func TestValidateAllowsSelectableURLTestAsDefault(t *testing.T) {
	cfg := Config{
		Version:      2,
		DefaultGroup: "AutoTest",
		Groups: []Group{{
			Name:       "AutoTest",
			Type:       "url-test",
			Selectable: true,
			URL:        "https://example.com",
			Interval:   300,
		}},
		Rules: []Rule{{Match: "MATCH", Policy: "AutoTest"}},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
