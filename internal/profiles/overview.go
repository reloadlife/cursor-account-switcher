package profiles

import (
	"time"

	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/platform"
)

type AccountOverview struct {
	ID       paths.AccountID
	Label    string
	Profile  *Profile
	Saved    bool
	IsActive bool
	IsLive   bool
}

type PlatformOverview struct {
	ID            platform.ID
	Name          string
	IsDefault     bool
	LiveID        string
	ActiveAccount *paths.AccountID
	Accounts      []AccountOverview
}

func WithPlatform(id platform.ID, fn func() error) error {
	prev := platform.Current()
	defer func() { _ = platform.SetCurrent(prev) }()
	if err := platform.SetCurrent(id); err != nil {
		return err
	}
	return fn()
}

func AllPlatformsOverview() ([]PlatformOverview, error) {
	global, err := LoadGlobalConfig()
	if err != nil {
		return nil, err
	}

	prev := platform.Current()
	defer func() { _ = platform.SetCurrent(prev) }()

	out := make([]PlatformOverview, 0, len(platform.List()))
	for _, p := range platform.List() {
		if err := platform.SetCurrent(p.ID); err != nil {
			continue
		}

		accounts, err := ListAccounts()
		if err != nil {
			return nil, err
		}

		liveID, _ := CurrentIdentifier()
		active := ActiveAccount()

		overview := PlatformOverview{
			ID:            p.ID,
			Name:          p.Name,
			IsDefault:     p.ID == global.ActivePlatform,
			LiveID:        liveID,
			ActiveAccount: active,
			Accounts:      make([]AccountOverview, 0, len(accounts)),
		}

		for _, account := range accounts {
			profile, _ := Load(account.ID)
			saved := profile != nil
			isActive := active != nil && *active == account.ID
			isLive := saved && profile.Email != nil && liveID != "" && *profile.Email == liveID

			overview.Accounts = append(overview.Accounts, AccountOverview{
				ID:       account.ID,
				Label:    account.Label,
				Profile:  profile,
				Saved:    saved,
				IsActive: isActive,
				IsLive:   isLive,
			})
		}

		out = append(out, overview)
	}

	return out, nil
}

func AccountLabelFor(platformID platform.ID, id paths.AccountID) string {
	var label string
	_ = WithPlatform(platformID, func() error {
		label = AccountLabel(id)
		return nil
	})
	if label == "" {
		return string(id)
	}
	return label
}

func FormatSavedAt(savedAt string) string {
	if t, err := time.Parse(time.RFC3339, savedAt); err == nil {
		return t.Local().Format("Jan 2 15:04")
	}
	return savedAt
}
