// Package config manages mctl config.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Group struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Selectable bool   `json:"selectable"`
	Filter     string `json:"filter,omitempty"`
	URL        string `json:"url,omitempty"`
	Interval   int    `json:"interval,omitempty"`
}

type Rule struct {
	Match   string   `json:"match"`
	Policy  string   `json:"policy"`
	Options []string `json:"options,omitempty"`
}

type Config struct {
	Version      int     `json:"version,omitempty"`
	DefaultGroup string  `json:"default_group,omitempty"`
	Groups       []Group `json:"groups,omitempty"`
	Rules        []Rule  `json:"rules,omitempty"`
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

func Dir() (string, error) {
	base, err := getXdgConfig()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "mihomo"), nil
}

func BasePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "base.yaml"), nil
}

func Load() (Config, error) {
	dir, err := Dir()
	if err != nil {
		return Config{}, err
	}
	configPath := filepath.Join(dir, "mctl.json")
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
	configDir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "mctl.json")
	return writeFileAtomic(configPath, data, 0o600)
}
