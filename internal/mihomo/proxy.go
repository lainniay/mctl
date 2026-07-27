package mihomo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *Client) Select(ctx context.Context, group, node string) error {
	body, err := json.Marshal(struct {
		Name string `json:"name"`
	}{Name: node})
	if err != nil {
		return err
	}
	endpoint := proxyEndpoint(c.addr, group)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("select node: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError("select node", resp)
	}
	return nil
}

func (c *Client) Delay(ctx context.Context, node, testURL string, timeout time.Duration) (int, error) {
	endpoint := proxyEndpoint(c.addr, node)
	endpoint.Path += "/delay"
	endpoint.RawPath += "/delay"
	query := endpoint.Query()
	query.Set("url", testURL)
	query.Set("timeout", strconv.FormatInt(timeout.Milliseconds(), 10))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, err
	}
	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("test node: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, responseError("test node", resp)
	}
	var result struct {
		Delay int `json:"delay"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode node delay: %w", err)
	}

	return result.Delay, nil
}

func proxyEndpoint(base *url.URL, name string) url.URL {
	endpoint := *base
	endpoint.Path = "/proxies/" + name
	endpoint.RawPath = "/proxies/" + url.PathEscape(name)
	endpoint.RawQuery = ""
	return endpoint
}

func (c *Client) authorize(req *http.Request) {
	if c.secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.secret)
	}
}

func responseError(action string, resp *http.Response) error {
	message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("%s: %s: %s", action, resp.Status, strings.TrimSpace(string(message)))
}
