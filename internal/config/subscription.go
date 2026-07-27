package config

import (
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"os"
)

var ErrSubExists = errors.New("subscription already exists")

type Sub struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type Subs struct {
	Items []Sub `json:"subs"`
}

func getSubPath() (string, error) {
	return statePath("subscriptions.json")
}

func LoadSubs() (Subs, error) {
	subPath, err := getSubPath()
	if err != nil {
		return Subs{}, err
	}
	data, err := os.ReadFile(subPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Subs{}, nil
		}
		return Subs{}, err
	}

	var subs Subs
	if err := json.Unmarshal(data, &subs); err != nil {
		return Subs{}, err
	}
	return subs, nil
}

func (s Subs) Save() error {
	path, err := getSubPath()
	if err != nil {
		return err
	}
	if err := ensureStateDir(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o600)
}

func (s *Subs) AddSub(name, url string) error {
	if name == "" {
		return fmt.Errorf("subscription name is required")
	}
	parsed, err := neturl.ParseRequestURI(url)
	if err != nil || parsed.Host == "" || parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid subscription URL")
	}
	for _, item := range s.Items {
		if item.Name == name || item.URL == url {
			return fmt.Errorf("%w: %s", ErrSubExists, name)
		}
	}
	s.Items = append(s.Items, Sub{Name: name, URL: url, Enabled: true})
	return nil
}

func (s *Subs) RemoveSub(name string) bool {
	for i, item := range s.Items {
		if item.Name == name {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Subs) SetSubEnabled(name string, enabled bool) bool {
	for i, item := range s.Items {
		if item.Name == name {
			s.Items[i].Enabled = enabled
			return true
		}
	}
	return false
}
