package app

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/process"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

type StepFn func(label string)

func SwitchTo(id paths.AccountID, onStep StepFn) error {
	step := func(label string) {
		if onStep != nil {
			onStep(label)
		}
	}

	if !profiles.Exists(id) {
		return fmt.Errorf(
			"profile %q not saved yet — log into that account in Cursor, then run: cursor-switch save %s",
			profiles.AccountLabel(id), id,
		)
	}

	active := profiles.ActiveAccount()
	currentEmail, _ := profiles.CurrentEmail()
	target, _ := profiles.Load(id)

	if target != nil && active != nil && *active == id {
		if target.Email != nil && currentEmail == *target.Email {
			step(fmt.Sprintf("Already on %s (%s)", profiles.AccountLabel(id), *target.Email))
			return nil
		}
	}

	step("Saving current session...")
	profiles.AutoSaveActive()

	step("Force quitting Cursor...")
	if err := process.ForceQuitCursor(); err != nil {
		return err
	}

	step("Restoring auth session...")
	restored, err := profiles.Restore(id)
	if err != nil {
		return err
	}

	email := profiles.AccountLabel(id)
	if restored.Email != nil {
		email = *restored.Email
	}

	step("Starting Cursor...")
	if err := process.StartCursor(); err != nil {
		return err
	}

	step(fmt.Sprintf("Switched to %s (%s)", profiles.AccountLabel(id), email))
	return nil
}
