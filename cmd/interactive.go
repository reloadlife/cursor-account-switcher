package cmd

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/reloadlife/cursor-account-switcher/internal/tui"
)

func handleMenuAction(result tui.MenuResult) error {
	switch result.Kind {
	case tui.ActionQuit, tui.ActionNone:
		return nil

	case tui.ActionStatus:
		fmt.Println(app.StatusText())
		return nil

	case tui.ActionSwitch:
		return tui.RunSwitch(result.AccountID)

	case tui.ActionSave:
		profile, err := app.SaveAs(result.AccountID, "")
		if err != nil {
			return err
		}
		email := "unknown"
		if profile.Email != nil {
			email = *profile.Email
		}
		fmt.Println(tui.FormatSaveSuccess(result.AccountID, email))
		return nil

	case tui.ActionAddAccount:
		id, err := tui.RunAddAccount()
		if err != nil {
			return err
		}
		if id != "" {
			fmt.Printf("Added account %s (%s)\n", app.AccountLabel(id), id)
		}
		return nil
	}

	return nil
}

func runInteractive() error {
	result, err := tui.RunMenu()
	if err != nil {
		return err
	}
	return handleMenuAction(result)
}
