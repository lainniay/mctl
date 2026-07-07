// Package config manages mctl config.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
)

var ErrSubExists = errors.New("subscription already exists")

type Sub struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type Config struct {
	Subs []Sub `json:"subs"`
}

func getXdgConfig() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" || !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return base, nil
}

func Load() (Config, error) {
	base, err := getXdgConfig()
	if err != nil {
		return Config{}, err
	}
	configPath := filepath.Join(base, "mihomo", "mctl.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Save() error {
	base, err := getXdgConfig()
	if err != nil {
		return err
	}
	configDir := filepath.Join(base, "mihomo")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "mctl.json")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return err
	}
	return nil
}

func (c *Config) AddSub(name, url string) error {
	for _, sub := range c.Subs {
		if sub.Name == name || sub.URL == url {
			return ErrSubExists
		}
	}
	c.Subs = append(c.Subs, Sub{Name: name, URL: url, Enabled: true})
	return nil
}

func (c *Config) RemoveSub(nameOrURL string) bool {
	for i, sub := range c.Subs {
		if sub.Name == nameOrURL || sub.URL == nameOrURL {
			c.Subs = slices.Delete(c.Subs, i, i+1)
			return true
		}
	}
	return false
}

func (c *Config) SetSubEnabled(nameOrURL string, enabled bool) bool {
	for i, sub := range c.Subs {
		if sub.Name == nameOrURL || sub.URL == nameOrURL {
			c.Subs[i].Enabled = enabled
			return true
		}
	}
	return false
}
