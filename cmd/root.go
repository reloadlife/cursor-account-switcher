package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cursor-switch",
	Short: "Switch between Cursor personal and work accounts",
	Long: `cursor-switch swaps saved Cursor auth sessions between personal and work accounts.

Force-quits Cursor, restores the selected profile, and restarts the app.`,
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
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(accountCmd)
}
