package app

import (
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

func SaveAs(id paths.AccountID, label string) (*profiles.Profile, error) {
	return profiles.SaveCurrentAs(id, profiles.SaveOptions{Label: label})
}
