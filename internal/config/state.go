package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type State struct {
	CurrentGroup string `json:"current_group"`
}

func getXdgState() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" || !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	return base, nil
}

func StateDir() (string, error) {
	base, err := getXdgState()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "mihomo"), nil
}

func statePath(name string) (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

func ensureStateDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.Chmod(dir, 0o700)
}

func BackupPath() (string, error) {
	return statePath("config.yaml.bak")
}

func LoadState() (State, error) {
	path, err := statePath("state.json")
	if err != nil {
		return State{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, nil
		}
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func (s State) Save() error {
	path, err := statePath("state.json")
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
