package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"
	appconfig "github.com/lainniay/mctl/internal/config"
	"github.com/lainniay/mctl/internal/mihomo"
	"github.com/spf13/cobra"
)

const defaultDelayURL = "https://www.gstatic.com/generate_204"

type delayTestPreset struct {
	Name        string
	URL         string
	Description string
}

var delayTestPresets = []delayTestPreset{
	{Name: "chatgpt", URL: "https://chatgpt.com/cdn-cgi/trace", Description: "ChatGPT connectivity"},
	{Name: "youtube", URL: "https://www.youtube.com/generate_204", Description: "YouTube connectivity"},
	{Name: "google", URL: defaultDelayURL, Description: "Google connectivity"},
	{Name: "github", URL: "https://github.com", Description: "GitHub connectivity"},
}

var (
	nodeTestURL     string
	nodeTestPreset  string
	nodeTestTimeout time.Duration
)

func init() {
	rootCommand.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeListCmd, nodeCurrentCmd, nodeUseCmd, nodeTestCmd)
	nodeTestCmd.Flags().StringVar(&nodeTestURL, "url", defaultDelayURL, "URL used for delay testing")
	nodeTestCmd.Flags().StringVar(&nodeTestPreset, "preset", "", "connectivity test preset (chatgpt, youtube, google, github)")
	nodeTestCmd.Flags().DurationVar(&nodeTestTimeout, "timeout", 5*time.Second, "delay test timeout")
	nodeTestCmd.MarkFlagsMutuallyExclusive("preset", "url")
	if err := nodeTestCmd.RegisterFlagCompletionFunc("preset", completeDelayTestPresets); err != nil {
		panic(err)
	}
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage mihomo nodes",
}

var nodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List nodes in the current group",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, group, err := currentRuntimeGroup(cmd.Context())
		if err != nil {
			return err
		}

		out := commandOutput(cmd)
		if out.pretty {
			rows := make([][]string, 0, len(group.All))
			styles := make([]*color.Color, 0, len(group.All))
			for _, node := range group.All {
				marker := ""
				style := out.cyan
				if node == group.Now {
					marker = iconCurrent
					style = out.green
				}
				rows = append(rows, []string{marker, node})
				styles = append(styles, style)
			}
			return out.table([]string{"", "NODE"}, rows, styles)
		}
		for _, node := range group.All {
			if _, err := fmt.Fprintf(out.w, "%s\t%s\n", out.marker(node == group.Now), out.name(node)); err != nil {
				return err
			}
		}
		return nil
	},
}

var nodeCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current group and node",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, group, err := currentRuntimeGroup(cmd.Context())
		if err != nil {
			return err
		}

		out := commandOutput(cmd)
		_, err = fmt.Fprintf(out.w, "%s\t%s\t%s\n", out.name(group.Name), out.detail(group.Type), out.name(group.Now))
		return err
	},
}

var nodeUseCmd = &cobra.Command{
	Use:               "use <node>",
	Short:             "Select a node in the current Selector group",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSelectorNodes,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, group, err := currentRuntimeGroup(cmd.Context())
		if err != nil {
			return err
		}

		if group.Type != "Selector" {
			return fmt.Errorf("group %q is %s and selects nodes automatically", group.Name, group.Type)
		}

		if !slices.Contains(group.All, args[0]) {
			return fmt.Errorf("node %q is not in group %q", args[0], group.Name)
		}

		if err := client.Select(cmd.Context(), group.Name, args[0]); err != nil {
			return err
		}

		return commandOutput(cmd).successf("current node: %s", args[0])
	},
}

var nodeTestCmd = &cobra.Command{
	Use:               "test <node>",
	Short:             "Test delay for a node in the current group",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeCurrentGroupNodes,
	RunE: func(cmd *cobra.Command, args []string) error {
		testURL, err := resolveDelayTestURL(nodeTestPreset, nodeTestURL, cmd.Flags().Changed("url"))
		if err != nil {
			return err
		}
		if nodeTestTimeout <= 0 {
			return fmt.Errorf("positive timeout is required")
		}

		client, group, err := currentRuntimeGroup(cmd.Context())
		if err != nil {
			return err
		}

		if !slices.Contains(group.All, args[0]) {
			return fmt.Errorf("node %q is not in group %q", args[0], group.Name)
		}

		delay, err := client.ProviderDelay(
			cmd.Context(), mihomo.ManagedProviderName, args[0], testURL, nodeTestTimeout,
		)
		if err != nil {
			return err
		}

		out := commandOutput(cmd)
		_, err = fmt.Fprintf(out.w, "%s\t%s\n", out.name(args[0]), out.detail(fmt.Sprintf("%dms", delay)))
		return err
	},
}

func resolveDelayTestURL(preset, testURL string, customURLSet bool) (string, error) {
	if preset == "" {
		if testURL == "" {
			return "", fmt.Errorf("test url is required")
		}
		return testURL, nil
	}
	if customURLSet {
		return "", fmt.Errorf("--preset and --url cannot be used together")
	}
	for _, candidate := range delayTestPresets {
		if strings.EqualFold(candidate.Name, preset) {
			return candidate.URL, nil
		}
	}
	names := make([]string, 0, len(delayTestPresets))
	for _, candidate := range delayTestPresets {
		names = append(names, candidate.Name)
	}
	return "", fmt.Errorf("unknown test preset %q (available: %s)", preset, strings.Join(names, ", "))
}

func completeDelayTestPresets(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	result := make([]string, 0, len(delayTestPresets))
	for _, preset := range delayTestPresets {
		if strings.HasPrefix(preset.Name, strings.ToLower(toComplete)) {
			result = append(result, preset.Name+"\t"+preset.Description)
		}
	}
	return result, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

func currentRuntimeGroup(ctx context.Context) (*mihomo.Client, mihomo.Group, error) {
	cfg, err := appconfig.Load()
	if err != nil {
		return nil, mihomo.Group{}, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, mihomo.Group{}, err
	}
	state, err := appconfig.LoadState()
	if err != nil {
		return nil, mihomo.Group{}, err
	}
	name := currentGroupName(cfg, state)
	controller, err := appconfig.LoadControllerFromEnv()
	if err != nil {
		return nil, mihomo.Group{}, err
	}

	client, err := mihomo.NewClient(controller.Address, controller.Secret)
	if err != nil {
		return nil, mihomo.Group{}, err
	}
	group, exists, err := client.Group(ctx, name)
	if err != nil {
		return nil, mihomo.Group{}, err
	}
	if !exists {
		return nil, mihomo.Group{}, fmt.Errorf("current group %q is missing from mihomo", name)
	}
	return client, group, nil
}

func completeSelectorNodes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeNodes(cmd, args, toComplete, true)
}

func completeCurrentGroupNodes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeNodes(cmd, args, toComplete, false)
}

func completeNodes(cmd *cobra.Command, args []string, toComplete string, selectorOnly bool) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
	defer cancel()
	_, group, err := currentRuntimeGroup(ctx)
	if err != nil || selectorOnly && group.Type != "Selector" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	result := make([]string, 0, len(group.All))
	for _, node := range group.All {
		if !strings.HasPrefix(node, toComplete) {
			continue
		}
		if node == group.Now {
			result = append(result, node+"\tcurrent")
		} else {
			result = append(result, node)
		}
	}
	return result, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}
