package cmd

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
	"github.com/spf13/cobra"
)

var materializeDest string

var materializeCmd = &cobra.Command{
	Use:   "materialize <account>",
	Short: "Write a profile's auth files under an isolated HOME directory",
	Long: `Materialize expands a saved profile into dest HOME without changing the global login.
Useful for agentsd parallel sessions:

  cursor-switch --platform grok materialize work --dest /tmp/grok-work-home
  HOME=/tmp/grok-work-home grok
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if materializeDest == "" {
			return fmt.Errorf("--dest is required")
		}
		id, err := profiles.ResolveAccount(args[0])
		if err != nil {
			return err
		}
		out, err := profiles.Materialize(id, materializeDest)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	materializeCmd.Flags().StringVar(&materializeDest, "dest", "", "destination HOME directory")
}
