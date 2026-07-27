package sub

import (
	"strings"
	"testing"
)

func TestFetchDoesNotExposeSubscriptionURL(t *testing.T) {
	const secret = "super-secret-token"
	_, err := Fetch("http://127.0.0.1:1/sub?token=" + secret)
	if err == nil {
		t.Fatal("Fetch() error = nil")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("Fetch() exposed subscription URL: %v", err)
	}
}
