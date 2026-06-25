package app

import (
	"fmt"
	"strings"

	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/profiles"
)

func ResolveAccount(ref string) (paths.AccountID, error) {
	return profiles.ResolveAccount(ref)
}

func ParseAccountID(arg string) (paths.AccountID, error) {
	id := paths.AccountID(strings.ToLower(strings.TrimSpace(arg)))
	if err := profiles.ValidateID(id); err != nil {
		return "", err
	}
	return id, nil
}

func AddAccount(id paths.AccountID, label string) error {
	return profiles.AddAccount(id, label)
}

func RegisterAccountFromLabel(label string) (paths.AccountID, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return "", fmt.Errorf("label is required")
	}
	id := paths.AccountID(profiles.SlugFromLabel(label))
	if profiles.AccountExists(id) {
		return "", fmt.Errorf("account %q already exists — pick a different label", id)
	}
	if err := profiles.AddAccount(id, label); err != nil {
		return "", err
	}
	return id, nil
}

func RemoveAccount(id paths.AccountID) error {
	return profiles.RemoveAccount(id)
}

func ListAccounts() ([]profiles.AccountDef, error) {
	return profiles.ListAccounts()
}
