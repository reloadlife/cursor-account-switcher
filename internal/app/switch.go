package app

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/platform"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

type StepFn func(label string)

func SwitchTo(id paths.AccountID, onStep StepFn) error {
	step := func(label string) {
		if onStep != nil {
			onStep(label)
		}
	}

	p, err := platform.CurrentPlatform()
	if err != nil {
		return err
	}

	if !profiles.Exists(id) {
		return fmt.Errorf(
			"profile %q not saved yet — log into %s, then run: cursor-switch save %s",
			profiles.AccountLabel(id), p.Name, id,
		)
	}

	active := profiles.ActiveAccount()
	currentID, _ := profiles.CurrentIdentifier()
	target, _ := profiles.Load(id)

	if target != nil && active != nil && *active == id {
		targetID := ""
		if target.Email != nil {
			targetID = *target.Email
		}
		if targetID != "" && currentID == targetID {
			step(fmt.Sprintf("Already on %s (%s)", profiles.AccountLabel(id), targetID))
			return nil
		}
	}

	step("Saving current session...")
	profiles.AutoSaveActive()

	if p.NeedsRestart() {
		step(fmt.Sprintf("Force quitting %s...", p.Name))
		if err := p.ForceQuit(); err != nil {
			return err
		}
	}

	step("Restoring auth session...")
	restored, err := profiles.Restore(id)
	if err != nil {
		return err
	}

	identifier := profiles.AccountLabel(id)
	if restored.Email != nil {
		identifier = *restored.Email
	}

	if p.NeedsRestart() {
		step(fmt.Sprintf("Starting %s...", p.Name))
		if err := p.Start(); err != nil {
			return err
		}
	}

	step(fmt.Sprintf("Switched to %s (%s)", profiles.AccountLabel(id), identifier))
	return nil
}

func PlatformName() string {
	p, err := platform.CurrentPlatform()
	if err != nil {
		return "Cursor"
	}
	return p.Name
}
