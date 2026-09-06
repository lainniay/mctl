package sub

import (
	"reflect"
	"testing"
)

func TestCleanMakesDistinctSSConnectionNamesUnique(t *testing.T) {
	var proxies []Proxy
	for i, raw := range []string{
		"ss://aes-256-gcm:one@example.com:443",
		"ss://aes-256-gcm:two@example.com:443",
		"ss://aes-256-gcm:three@example.com:443#example.com:443-2",
		"ss://aes-256-gcm:four@example.com:443",
	} {
		p, err := ParseURL(raw)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		proxies = append(proxies, p)
	}
	got := Clean(proxies)
	var names []string
	for _, p := range got {
		names = append(names, p.Name)
	}
	want := []string{"example.com:443", "example.com:443-3", "example.com:443-2", "example.com:443-4"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("names=%v, want %v", names, want)
	}
}
