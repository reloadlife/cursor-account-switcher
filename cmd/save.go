package cmd

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/reloadlife/cursor-account-switcher/internal/tui"
	"github.com/spf13/cobra"
)

var saveLabel string

var saveCmd = &cobra.Command{
	Use:   "save [account]",
	Short: "Save the current Cursor session as a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := app.ResolveAccount(args[0])
		if err != nil {
			id, err = app.ParseAccountID(args[0])
			if err != nil {
				return err
			}
		}

		profile, err := app.SaveAs(id, saveLabel)
		if err != nil {
			return err
		}

		email := "unknown"
		if profile.Email != nil {
			email = *profile.Email
		}

		fmt.Println(tui.FormatSaveSuccess(id, email))
		return nil
	},
}

func init() {
	saveCmd.Flags().StringVar(&saveLabel, "label", "", "display label (required when creating a new account)")
}
