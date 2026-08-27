package app

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/platform"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

type StepFn func(label string)

type SwitchOptions struct {
	NoStart bool
}

type appStarter interface {
	NeedsRestart() bool
	Start() error
	AppName() string
}

func startAppIfNeeded(p appStarter, opts SwitchOptions, step StepFn) error {
	if !p.NeedsRestart() || opts.NoStart {
		return nil
	}
	if step != nil {
		step(fmt.Sprintf("Starting %s...", p.AppName()))
	}
	return p.Start()
}

type platformStarter struct {
	p *platform.Platform
}

func (ps platformStarter) NeedsRestart() bool { return ps.p.NeedsRestart() }
func (ps platformStarter) Start() error       { return ps.p.Start() }
func (ps platformStarter) AppName() string    { return ps.p.Name }

func SwitchTo(id paths.AccountID, opts SwitchOptions, onStep StepFn) error {
	step := func(label string) {
		if onStep != nil {
			onStep(label)
		}
	}

	p, err := platform.CurrentPlatform()
	if err != nil {
		return err
	}

	// Empty slot (registered, no saved credentials): clear live auth so user can sign in.
	if !profiles.Exists(id) {
		if !profiles.AccountExists(id) {
			return fmt.Errorf(
				"unknown account %q — run: cursor-switch account add %s --label \"…\"",
				id, id,
			)
		}
		step("Saving current session...")
		profiles.AutoSaveActive()

		if p.NeedsRestart() {
			step(fmt.Sprintf("Force quitting %s...", p.Name))
			if err := p.ForceQuit(); err != nil {
				return err
			}
		}

		step(fmt.Sprintf("Clearing live auth for empty profile %s...", profiles.AccountLabel(id)))
		if err := profiles.ActivateEmpty(id); err != nil {
			return err
		}

		if err := startAppIfNeeded(platformStarter{p}, opts, step); err != nil {
			return err
		}

		step(fmt.Sprintf(
			"Empty profile %s is active — sign into %s, then: cursor-switch save %s",
			profiles.AccountLabel(id), p.Name, id,
		))
		return nil
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

	if err := startAppIfNeeded(platformStarter{p}, opts, step); err != nil {
		return err
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
