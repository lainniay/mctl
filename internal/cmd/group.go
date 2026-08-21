package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	appconfig "github.com/lainniay/mctl/internal/config"
	"github.com/lainniay/mctl/internal/mihomo"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(groupCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupChangeCmd)
}

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage mihomo groups",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all selectable groups",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := appconfig.Load()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}

		controller, err := appconfig.LoadControllerFromEnv()
		if err != nil {
			return err
		}
		client, err := mihomo.NewClient(controller.Address, controller.Secret)
		if err != nil {
			return err
		}

		runtimeGroups, err := client.Groups(cmd.Context())
		if err != nil {
			return err
		}

		groups, err := selectableGroups(cfg, runtimeGroups)
		if err != nil {
			return err
		}

		state, err := appconfig.LoadState()
		if err != nil {
			return err
		}

		current := currentGroupName(cfg, state)
		out := commandOutput(cmd)
		if out.pretty {
			rows := make([][]string, 0, len(groups))
			styles := make([]*color.Color, 0, len(groups))
			for _, group := range groups {
				marker := ""
				style := out.cyan
				if group.Name == current {
					marker = iconCurrent
					style = out.green
				} else if !group.Alive {
					style = out.red
				}
				alive := iconFailure
				if group.Alive {
					alive = iconSuccess
				}
				rows = append(rows, []string{marker, group.Name, group.Type, group.Now, fmt.Sprint(len(group.All)), alive})
				styles = append(styles, style)
			}
			return out.table([]string{"", "GROUP", "TYPE", "NODE", "NODES", "ALIVE"}, rows, styles)
		}
		for _, group := range groups {
			if _, err := fmt.Fprintf(
				out.w,
				"%s\t%s\t%s\t%s\t%s\t%s\n",
				out.marker(group.Name == current),
				out.name(group.Name),
				out.detail(group.Type),
				out.name(group.Now),
				out.detail(len(group.All)),
				out.status(group.Alive),
			); err != nil {
				return err
			}
		}
		return nil
	},
}

var groupChangeCmd = &cobra.Command{
	Use:               "change <group>",
	Short:             "Change the group used by node commands",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSelectableGroups,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := appconfig.Load()
		if err != nil {
			return err
		}
		if err = cfg.Validate(); err != nil {
			return err
		}
		selectable := false
		for _, group := range cfg.Groups {
			if group.Name == name {
				selectable = group.Selectable
				break
			}
		}
		if !selectable {
			return fmt.Errorf("group %q is not selectable", name)
		}

		controller, err := appconfig.LoadControllerFromEnv()
		if err != nil {
			return err
		}
		client, err := mihomo.NewClient(controller.Address, controller.Secret)
		if err != nil {
			return err
		}
		_, exists, err := client.Group(cmd.Context(), name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("group %q is missing from mihomo", name)
		}

		state, err := appconfig.LoadState()
		if err != nil {
			return err
		}
		state.CurrentGroup = name
		if err := state.Save(); err != nil {
			return err
		}

		return commandOutput(cmd).successf("current group: %s", name)
	},
}

func selectableGroups(cfg appconfig.Config, runtime []mihomo.Group) ([]mihomo.Group, error) {
	byName := make(map[string]mihomo.Group, len(runtime))
	for _, group := range runtime {
		byName[group.Name] = group
	}

	var result []mihomo.Group
	for _, declared := range cfg.Groups {
		if !declared.Selectable {
			continue
		}

		group, exists := byName[declared.Name]
		if !exists {
			return nil, fmt.Errorf("selectable group %q is missing from mihomo", declared.Name)
		}
		result = append(result, group)
	}
	return result, nil
}

func completeSelectableGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := appconfig.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var result []string
	for _, group := range cfg.Groups {
		if !group.Selectable {
			continue
		}
		if !strings.HasPrefix(group.Name, toComplete) {
			continue
		}
		result = append(result, group.Name+"\t"+group.Type)
	}
	return result,
		cobra.ShellCompDirectiveNoFileComp |
			cobra.ShellCompDirectiveKeepOrder
}

func currentGroupName(cfg appconfig.Config, state appconfig.State) string {
	if state.CurrentGroup != "" {
		for _, group := range cfg.Groups {
			if group.Name == state.CurrentGroup && group.Selectable {
				return state.CurrentGroup
			}
		}
	}
	return cfg.DefaultGroup
}
