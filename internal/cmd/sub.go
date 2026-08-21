package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/lainniay/mctl/internal/config"
	"github.com/lainniay/mctl/internal/sub"
	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Manage subscription links",
}

var subAddCmd = &cobra.Command{
	Use:     "add <name>",
	Short:   "Add subscription link",
	Example: `  MCTL_SUB_URL="$(pbpaste)" mctl sub add primary`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := os.Getenv("MCTL_SUB_URL")
		if url == "" {
			return fmt.Errorf("MCTL_SUB_URL is required")
		}
		subs, err := config.LoadSubs()
		if err != nil {
			return err
		}
		if err := subs.AddSub(args[0], url); err != nil {
			if errors.Is(err, config.ErrSubExists) {
				return fmt.Errorf("subscription already exists: %s", args[0])
			}
			return err
		}
		return subs.Save()
	},
}

var subRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove subscription link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subs, err := config.LoadSubs()
		if err != nil {
			return err
		}
		if !subs.RemoveSub(args[0]) {
			return fmt.Errorf("subscription not found: %s", args[0])
		}
		return subs.Save()
	},
}

var subListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all subscription links",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		subs, err := config.LoadSubs()
		if err != nil {
			return err
		}
		out := commandOutput(cmd)
		if out.pretty {
			rows := make([][]string, 0, len(subs.Items))
			styles := make([]*color.Color, 0, len(subs.Items))
			for _, sub := range subs.Items {
				status := iconFailure
				style := out.red
				if sub.Enabled {
					status = iconSuccess
					style = out.green
				}
				rows = append(rows, []string{sub.Name, status})
				styles = append(styles, style)
			}
			return out.table([]string{"NAME", "STATUS"}, rows, styles)
		}
		for _, sub := range subs.Items {
			if _, err := fmt.Fprintf(out.w, "%s\t%s\n", out.name(sub.Name), out.status(sub.Enabled)); err != nil {
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
		return commandOutput(cmd).successf("wrote %d proxies to %s", count, path)
	},
}

func runSubUpdate() (int, string, error) {
	subs, err := config.LoadSubs()
	if err != nil {
		return 0, "", err
	}

	var proxies []sub.Proxy
	for _, source := range subs.Items {
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
	if err := writeAtomic(path, provider, 0o644, 0o755); err != nil {
		return 0, "", err
	}
	return len(proxies), path, nil
}

func providerPath() (string, error) {
	dir, err := config.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "providers", "nodes.yaml"), nil
}

var subEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setSubEnabled(args[0], true)
	},
}

var subDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setSubEnabled(args[0], false)
	},
}

func setSubEnabled(name string, enabled bool) error {
	subs, err := config.LoadSubs()
	if err != nil {
		return err
	}
	if !subs.SetSubEnabled(name, enabled) {
		return fmt.Errorf("subscription not found: %s", name)
	}
	return subs.Save()
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
