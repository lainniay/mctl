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
		switch r.Method {
		case http.MethodPut:
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
		case http.MethodGet:
			if r.URL.EscapedPath() != "/proxies/Japan%2001%20Pro/delay" {
				t.Errorf("delay path = %q", r.URL.EscapedPath())
			}
			if r.URL.Query().Get("timeout") != "5000" || r.URL.Query().Get("url") != defaultTestURL {
				t.Errorf("delay query = %v", r.URL.Query())
			}
			_, _ = fmt.Fprint(w, `{"delay":83}`)
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
}

const defaultTestURL = "https://www.gstatic.com/generate_204"
