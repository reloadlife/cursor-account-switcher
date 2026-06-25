package cmd

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
	"github.com/spf13/cobra"
)

var accountLabel string

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage registered accounts",
}

var accountAddCmd = &cobra.Command{
	Use:   "add [id]",
	Short: "Register a new account",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		label := accountLabel
		if label == "" {
			return fmt.Errorf("--label is required")
		}

		var id paths.AccountID
		if len(args) == 1 {
			parsed, err := app.ParseAccountID(args[0])
			if err != nil {
				return err
			}
			id = parsed
		} else {
			id = paths.AccountID(profiles.SlugFromLabel(label))
		}

		if err := app.AddAccount(id, label); err != nil {
			return err
		}

		fmt.Printf("Added account %s (%s)\n", label, id)
		return nil
	},
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		accounts, err := app.ListAccounts()
		if err != nil {
			return err
		}
		for _, a := range accounts {
			saved := ""
			if profiles.ProfileSaved(a.ID) {
				saved = " · saved"
			}
			fmt.Printf("  %s (%s)%s\n", a.Label, a.ID, saved)
		}
		return nil
	},
}

var accountRemoveCmd = &cobra.Command{
	Use:   "remove [account]",
	Short: "Remove an account and its saved profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := app.ResolveAccount(args[0])
		if err != nil {
			return err
		}
		if err := app.RemoveAccount(id); err != nil {
			return err
		}
		fmt.Printf("Removed account %s\n", profiles.AccountLabel(id))
		return nil
	},
}

func init() {
	accountAddCmd.Flags().StringVar(&accountLabel, "label", "", "display label for the account")
	accountCmd.AddCommand(accountAddCmd, accountListCmd, accountRemoveCmd)
}
