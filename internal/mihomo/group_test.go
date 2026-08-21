package mihomo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGroupsDecodesRuntimeGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"proxies":[{"name":"NodeSelection","type":"Selector","now":"Japan 01 Pro","all":["Japan 01 Pro","Japan 02 Pro"],"alive":true,"hidden":false}]}`)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	groups, err := client.Groups(context.Background())
	if err != nil {
		t.Fatalf("Groups() error = %v", err)
	}
	if len(groups) != 1 || groups[0].Name != "NodeSelection" || groups[0].Now != "Japan 01 Pro" {
		t.Fatalf("Groups() = %+v", groups)
	}
	if !groups[0].Alive || len(groups[0].All) != 2 {
		t.Fatalf("Groups() = %+v", groups)
	}
}

func TestGroupGetsOnlyRequestedGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.EscapedPath() != "/proxies/Node%2FSelection" {
			t.Errorf("request = %s %s", r.Method, r.URL.EscapedPath())
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization = %q", got)
			http.Error(w, "unexpected authorization", http.StatusUnauthorized)
			return
		}
		_, _ = fmt.Fprint(w, `{"name":"Node/Selection","type":"Selector","now":"Japan 01","all":["Japan 01","Japan 02"],"alive":true}`)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	group, exists, err := client.Group(context.Background(), "Node/Selection")
	if err != nil {
		t.Fatalf("Group() error = %v", err)
	}
	if !exists || group.Name != "Node/Selection" || group.Now != "Japan 01" || len(group.All) != 2 {
		t.Fatalf("Group() = %+v, %t", group, exists)
	}
}
