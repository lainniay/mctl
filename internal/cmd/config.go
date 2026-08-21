package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appconfig "github.com/lainniay/mctl/internal/config"
	"github.com/lainniay/mctl/internal/mihomo"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage generated mihomo configuration",
}

var configRenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render the generated mihomo configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := renderConfig()
		if err != nil {
			return err
		}
		_, err = cmd.OutOrStdout().Write(data)
		return err
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the generated configuration with mihomo",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := runConfigValidate(ctx); err != nil {
			return err
		}
		return commandOutput(cmd).successf("configuration is valid")
	},
}

var configApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Validate, install, and reload the generated configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		path, err := runConfigApply(ctx)
		if err != nil {
			return err
		}
		return commandOutput(cmd).successf("applied configuration to %s", path)
	},
}

func renderConfig() ([]byte, string, error) {
	cfg, err := appconfig.Load()
	if err != nil {
		return nil, "", err
	}

	controller, err := appconfig.LoadControllerFromEnv()
	if err != nil {
		return nil, "", err
	}

	basePath, err := appconfig.BasePath()
	if err != nil {
		return nil, "", err
	}

	base, err := os.ReadFile(basePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("base config not found at %s", basePath)
		}
		return nil, "", fmt.Errorf("read base config %s: %w", basePath, err)
	}

	home, err := appconfig.StateDir()
	if err != nil {
		return nil, "", err
	}

	data, err := mihomo.RenderConfig(base, cfg, controller)

	return data, home, err
}

func runConfigValidate(ctx context.Context) error {
	data, home, err := renderConfig()
	if err != nil {
		return err
	}
	if err := checkProvider(home); err != nil {
		return err
	}
	return mihomo.ValidateConfig(ctx, data, home)
}

func checkProvider(home string) error {
	providerPath := filepath.Join(home, "providers", "nodes.yaml")
	if _, err := os.Stat(providerPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("provider not found at %s; run mctl sub update", providerPath)
		}
		return fmt.Errorf("check provider %s: %w", providerPath, err)
	}
	return nil
}

func runConfigApply(ctx context.Context) (string, error) {
	data, home, err := renderConfig()
	if err != nil {
		return "", err
	}
	if err := checkProvider(home); err != nil {
		return "", err
	}
	if err := mihomo.ValidateConfig(ctx, data, home); err != nil {
		return "", err
	}

	target := filepath.Join(home, "config.yaml")
	old, err := os.ReadFile(target)
	hadOld := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read active config %s: %w", target, err)
	}
	client, clientErr := mihomo.NewClientFromMihomoConfig(old)
	if !hadOld || clientErr != nil {
		client, clientErr = mihomo.NewClientFromMihomoConfig(data)
	}
	if clientErr != nil {
		return "", clientErr
	}

	if hadOld {
		backup, err := appconfig.BackupPath()
		if err != nil {
			return "", err
		}
		if err := writeAtomic(backup, old, 0o600, 0o700); err != nil {
			return "", fmt.Errorf("back up active config: %w", err)
		}
	}
	if err := writeAtomic(target, data, 0o600, 0o700); err != nil {
		return "", fmt.Errorf("install config: %w", err)
	}
	if err := client.ReloadConfig(ctx, target); err != nil {
		rollbackErr := rollbackConfig(ctx, client, target, old, hadOld)
		if rollbackErr != nil {
			return "", errors.Join(err, rollbackErr)
		}
		return "", err
	}
	return target, nil
}

func writeAtomic(path string, data []byte, mode, dirMode os.FileMode) (err error) {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, dirMode); err != nil {
		return err
	}
	if err := os.Chmod(dirPath, dirMode); err != nil {
		return err
	}
	file, err := os.CreateTemp(dirPath, ".mctl-*")
	if err != nil {
		return err
	}
	name := file.Name()
	closed := false
	renamed := false

	// Protection
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

	// Rename is a Atomic Operation, so write a temp file, then rename it is atomic
	if err := os.Rename(name, path); err != nil {
		return err
	}
	renamed = true
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func rollbackConfig(ctx context.Context, client *mihomo.Client, target string, old []byte, hadOld bool) error {
	if !hadOld {
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove failed config: %w", err)
		}
		return nil
	}
	if err := writeAtomic(target, old, 0o600, 0o700); err != nil {
		return fmt.Errorf("restore previous config: %w", err)
	}
	if err := client.ReloadConfig(ctx, target); err != nil {
		return fmt.Errorf("reload restored config: %w", err)
	}
	return nil
}

func init() {
	rootCommand.AddCommand(configCmd)
	configCmd.AddCommand(configRenderCmd, configValidateCmd, configApplyCmd)
}
