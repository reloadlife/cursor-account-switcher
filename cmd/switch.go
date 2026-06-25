package cmd

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/reloadlife/cursor-account-switcher/internal/tui"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [account]",
	Short: "Switch to a saved account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := app.ResolveAccount(args[0])
		if err != nil {
			return err
		}

		plain, _ := cmd.Flags().GetBool("plain")
		if plain {
			return app.SwitchTo(id, func(label string) {
				fmt.Println(label)
			})
		}

		return tui.RunSwitch(id)
	},
}

func init() {
	switchCmd.Flags().Bool("plain", false, "plain text output without spinner")
}
