package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lainniay/mctl/internal/config"
	"github.com/lainniay/mctl/internal/sub"
	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Manage subscription links",
}

var subAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add subscription link",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.AddSub(args[0], args[1]); err != nil {
			if errors.Is(err, config.ErrSubExists) {
				return fmt.Errorf("subscription already exists: %s", args[0])
			}
			return err
		}
		return cfg.Save()
	},
}

var subRemoveCmd = &cobra.Command{
	Use:   "remove <name-or-url>",
	Short: "Remove subscription link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if !cfg.RemoveSub(args[0]) {
			return fmt.Errorf("subscription not found: %s", args[0])
		}
		return cfg.Save()
	},
}

var subListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all subscription links",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		for _, sub := range cfg.Subs {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%t\n", sub.Name, sub.URL, sub.Enabled); err != nil {
				return err
			}
		}
		return nil
	},
}

var subUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local mihomo provider from enabled subscriptions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		count, path, err := runSubUpdate()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "wrote %d proxies to %s\n", count, path)
		return err
	},
}

func runSubUpdate() (int, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return 0, "", err
	}

	var proxies []sub.Proxy
	for _, source := range cfg.Subs {
		if !source.Enabled {
			continue
		}
		body, err := sub.Fetch(source.URL)
		if err != nil {
			return 0, "", fmt.Errorf("fetch %s: %w", source.Name, err)
		}
		items, err := sub.Parse(body)
		if err != nil {
			return 0, "", fmt.Errorf("parse %s: %w", source.Name, err)
		}
		proxies = append(proxies, items...)
	}

	proxies = sub.Clean(proxies)
	provider, err := sub.RenderProvider(proxies)
	if err != nil {
		return 0, "", err
	}

	path, err := providerPath()
	if err != nil {
		return 0, "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, "", err
	}
	if err := os.WriteFile(path, provider, 0o644); err != nil {
		return 0, "", err
	}
	return len(proxies), path, nil
}

func providerPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" || !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "mihomo", "providers", "nodes.yaml"), nil
}

var subEnableCmd = &cobra.Command{
	Use:   "enable <name-or-url>",
	Short: "Enable subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setSubEnabled(args[0], true)
	},
}

var subDisableCmd = &cobra.Command{
	Use:   "disable <name-or-url>",
	Short: "Disable subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setSubEnabled(args[0], false)
	},
}

func setSubEnabled(nameOrURL string, enabled bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !cfg.SetSubEnabled(nameOrURL, enabled) {
		return fmt.Errorf("subscription not found: %s", nameOrURL)
	}
	return cfg.Save()
}

func init() {
	rootCommand.AddCommand(subCmd)

	subCmd.AddCommand(subAddCmd)
	subCmd.AddCommand(subRemoveCmd)
	subCmd.AddCommand(subListCmd)
	subCmd.AddCommand(subUpdateCmd)
	subCmd.AddCommand(subEnableCmd)
	subCmd.AddCommand(subDisableCmd)
}
