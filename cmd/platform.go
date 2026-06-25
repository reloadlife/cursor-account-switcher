package cmd

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/platform"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
	"github.com/spf13/cobra"
)

var platformCmd = &cobra.Command{
	Use:     "platform",
	Aliases: []string{"p", "platforms", "plat"},
	Short:   "Manage target platforms",
}

var platformListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List supported platforms",
	RunE: func(cmd *cobra.Command, args []string) error {
		active, _ := profiles.LoadGlobalConfig()
		for _, p := range platform.List() {
			marker := " "
			if p.ID == active.ActivePlatform {
				marker = "*"
			}
			restart := "no restart"
			if p.NeedsRestart() {
				restart = "restarts app"
			}
			fmt.Printf("%s %-8s  %-14s  %s\n", marker, p.ID, p.Name, restart)
			fmt.Printf("           %s\n", p.Description)
		}
		fmt.Println("\n* = default platform")
		return nil
	},
}

var platformUseCmd = &cobra.Command{
	Use:     "use [platform]",
	Aliases: []string{"u"},
	Short:   "Set the default platform",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := platform.ID(args[0])
		if err := profiles.SetActivePlatform(id); err != nil {
			return err
		}
		p, _ := platform.Get(id)
		fmt.Printf("Default platform set to %s\n", p.Name)
		return nil
	},
}

func init() {
	platformCmd.AddCommand(platformListCmd, platformUseCmd)
}
