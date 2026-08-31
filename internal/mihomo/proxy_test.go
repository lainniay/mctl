package mihomo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientSelectAndDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization = %q", got)
		}
		switch {
		case r.Method == http.MethodPut:
			if r.URL.EscapedPath() != "/proxies/Group%2FOne" {
				t.Errorf("select path = %q", r.URL.EscapedPath())
			}
			var body struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name != "Japan 01 Pro" {
				t.Errorf("select body = %+v, err = %v", body, err)
			}
			w.WriteHeader(http.StatusNoContent)
		case r.URL.EscapedPath() == "/proxies/Japan%2001%20Pro/delay":
			assertDelayQuery(t, r)
			_, _ = fmt.Fprint(w, `{"delay":83}`)
		case r.URL.EscapedPath() == "/providers/proxies/Mctl%2FNodes/Japan%2002%20Pro/healthcheck":
			assertDelayQuery(t, r)
			_, _ = fmt.Fprint(w, `{"delay":91}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.EscapedPath())
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Select(context.Background(), "Group/One", "Japan 01 Pro"); err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	delay, err := client.Delay(context.Background(), "Japan 01 Pro", defaultTestURL, 5*time.Second)
	if err != nil {
		t.Fatalf("Delay() error = %v", err)
	}
	if delay != 83 {
		t.Fatalf("Delay() = %d, want 83", delay)
	}

	delay, err = client.ProviderDelay(
		context.Background(), "Mctl/Nodes", "Japan 02 Pro", defaultTestURL, 5*time.Second,
	)
	if err != nil {
		t.Fatalf("ProviderDelay() error = %v", err)
	}
	if delay != 91 {
		t.Fatalf("ProviderDelay() = %d, want 91", delay)
	}
}

func assertDelayQuery(t *testing.T, r *http.Request) {
	t.Helper()
	if r.URL.Query().Get("timeout") != "5000" || r.URL.Query().Get("url") != defaultTestURL {
		t.Errorf("delay query = %v", r.URL.Query())
	}
}

const defaultTestURL = "https://www.gstatic.com/generate_204"
