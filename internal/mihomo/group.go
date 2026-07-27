package mihomo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Group struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Now    string   `json:"now"`
	All    []string `json:"all"`
	Alive  bool     `json:"alive"`
	Hidden bool     `json:"hidden"`
}

type groupResponse struct {
	Proxies []Group `json:"proxies"`
}

func (c *Client) Groups(ctx context.Context) ([]Group, error) {
	endpoint := *c.addr
	endpoint.Path = "/group"
	endpoint.RawQuery = ""

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf(
			"get groups: %s: %s",
			resp.Status,
			strings.TrimSpace(string(message)),
		)
	}

	var result groupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode groups: %w", err)
	}

	return result.Proxies, nil
}
