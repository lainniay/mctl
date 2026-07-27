// Package mihomo talks to the mihomo external controller.
package mihomo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.yaml.in/yaml/v4"
)

type Client struct {
	addr   *url.URL
	secret string
	http   *http.Client
}

func NewClient(addr, secret string) (*Client, error) {
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}

	base, err := url.Parse(addr)
	if err != nil || base.Host == "" {
		return nil, fmt.Errorf("invalid external-controller address: %s", addr)
	}

	if base.Scheme != "http" && base.Scheme != "https" {
		return nil, fmt.Errorf("unsupported external-controller scheme: %s", base.Scheme)
	}

	if base.Port() == "" {
		return nil, fmt.Errorf("external-controller port is required: %s", addr)
	}

	host := base.Hostname()
	if host == "0.0.0.0" || host == "::" {
		if host == "0.0.0.0" {
			host = "127.0.0.1"
		} else {
			host = "::1"
		}
		base.Host = net.JoinHostPort(host, base.Port())
	}
	base.Path = ""
	base.RawPath = ""
	base.RawQuery = ""
	base.Fragment = ""
	return &Client{addr: base, secret: secret, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func NewClientFromMihomoConfig(data []byte) (*Client, error) {
	var cfg struct {
		Addr   string `yaml:"external-controller"`
		Secret string `yaml:"secret"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("read controller config: %w", err)
	}

	if cfg.Addr == "" {
		return nil, fmt.Errorf("external-controller is required")
	}

	return NewClient(cfg.Addr, cfg.Secret)
}

func (c *Client) ReloadConfig(ctx context.Context, path string) error {
	// create Json: {"path" : "path"} for Http Body
	body, err := json.Marshal(struct {
		Path string `json:"path"`
	}{Path: path})
	if err != nil {
		return err
	}

	// endpoint change to base/configs?force=true
	endpoint := *c.addr
	endpoint.Path = "/configs"
	query := endpoint.Query()
	query.Set("force", "true")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("reload config: %s: %s", resp.Status, strings.TrimSpace(string(message)))
	}
	return nil
}
