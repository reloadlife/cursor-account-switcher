package app

import (
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

func AccountLabel(id paths.AccountID) string {
	return profiles.AccountLabel(id)
}
