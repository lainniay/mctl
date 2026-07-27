package config

import (
	"os"
	"path/filepath"
)

// writeFileAtomic: rename a file is atomic, so write a temp file, rename it is atomic
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".mctl-*")
	if err != nil {
		return err
	}
	name := file.Name()
	closed := false
	renamed := false

	defer func() {
		if !closed {
			_ = file.Close()
		}
		if !renamed {
			_ = os.Remove(name)
		}
	}()

	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	closed = true

	if err := os.Rename(name, path); err != nil {
		return err
	}
	renamed = true

	// Sync the rename operation, ensure the rename operation is write into disk
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}

	defer func() {
		_ = dir.Close()
	}()

	return dir.Sync()
}
