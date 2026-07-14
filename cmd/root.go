package cmd

import (
	"fmt"
	"os"

	"github.com/reloadlife/cursor-account-switcher/internal/platform"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
	"github.com/spf13/cobra"
)

var platformFlag string

var rootCmd = &cobra.Command{
	Use:   "cursor-switch",
	Short: "Switch between saved accounts across AI coding platforms",
	Long: `cursor-switch swaps saved auth sessions between personal and work accounts.

Supports Cursor, Claude Code, Codex, Grok, VS Code / GitHub Copilot, and more.
Use --platform to target a specific tool, or "cursor-switch platform use" to set a default.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := profiles.InitPlatform(); err != nil {
			return err
		}
		if platformFlag != "" {
			return platform.SetCurrent(platform.ID(platformFlag))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractive()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&platformFlag, "platform", "", "platform: cursor, claude, codex, grok, vscode")
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(platformCmd)
	rootCmd.AddCommand(materializeCmd)
}
