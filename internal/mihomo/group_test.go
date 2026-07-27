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
