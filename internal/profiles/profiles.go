package profiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/reloadlife/cursor-account-switcher/internal/database"
	"github.com/reloadlife/cursor-account-switcher/internal/paths"
)

type AccountDef struct {
	ID    paths.AccountID `json:"id"`
	Label string          `json:"label"`
}

type Profile struct {
	ID       paths.AccountID   `json:"id"`
	Label    string            `json:"label"`
	Email    *string           `json:"email"`
	SavedAt  string            `json:"savedAt"`
	AuthKeys map[string]string `json:"authKeys"`
}

type Config struct {
	ActiveAccount *paths.AccountID `json:"activeAccount"`
	Accounts      []AccountDef     `json:"accounts"`
}

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func DefaultAccounts() []AccountDef {
	return []AccountDef{
		{ID: "personal", Label: "Personal"},
		{ID: "work", Label: "Work"},
	}
}

func SlugFromLabel(label string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(label)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "account"
	}
	return slug
}

func ValidateID(id paths.AccountID) error {
	if !idPattern.MatchString(string(id)) {
		return fmt.Errorf("invalid account id %q — use lowercase letters, numbers, and hyphens", id)
	}
	return nil
}

func ensureDataDir() error {
	if err := os.MkdirAll(filepath.Join(paths.DataDir(), "profiles"), 0o700); err != nil {
		return err
	}
	return os.MkdirAll(paths.DataDir(), 0o700)
}

func loadConfig() (Config, error) {
	data, err := os.ReadFile(paths.ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Config{Accounts: DefaultAccounts()}
			if writeErr := saveConfig(cfg); writeErr != nil {
				return cfg, writeErr
			}
			return cfg, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if len(cfg.Accounts) == 0 {
		cfg.Accounts = DefaultAccounts()
		_ = saveConfig(cfg)
	}

	return cfg, nil
}

func saveConfig(cfg Config) error {
	if err := ensureDataDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.ConfigPath(), data, 0o600)
}

func ListAccounts() ([]AccountDef, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return cfg.Accounts, nil
}

func AccountLabel(id paths.AccountID) string {
	cfg, err := loadConfig()
	if err != nil {
		return string(id)
	}
	for _, a := range cfg.Accounts {
		if a.ID == id {
			return a.Label
		}
	}
	if p, _ := Load(id); p != nil && p.Label != "" {
		return p.Label
	}
	return string(id)
}

func AccountExists(id paths.AccountID) bool {
	cfg, err := loadConfig()
	if err != nil {
		return false
	}
	for _, a := range cfg.Accounts {
		if a.ID == id {
			return true
		}
	}
	return false
}

func ResolveAccount(ref string) (paths.AccountID, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("account name required")
	}

	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}

	lower := strings.ToLower(ref)
	for _, a := range cfg.Accounts {
		if string(a.ID) == lower {
			return a.ID, nil
		}
	}

	for _, a := range cfg.Accounts {
		if strings.EqualFold(a.Label, ref) {
			return a.ID, nil
		}
	}

	return "", fmt.Errorf("unknown account %q — run cursor-switch account list", ref)
}

func RegisterAccount(id paths.AccountID, label string) error {
	if err := ValidateID(id); err != nil {
		return err
	}
	label = strings.TrimSpace(label)
	if label == "" {
		return fmt.Errorf("label is required")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	for i, a := range cfg.Accounts {
		if a.ID == id {
			cfg.Accounts[i].Label = label
			return saveConfig(cfg)
		}
	}

	cfg.Accounts = append(cfg.Accounts, AccountDef{ID: id, Label: label})
	return saveConfig(cfg)
}

func AddAccount(id paths.AccountID, label string) error {
	if AccountExists(id) {
		return fmt.Errorf("account %q already exists", id)
	}
	return RegisterAccount(id, label)
}

func RemoveAccount(id paths.AccountID) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	found := false
	next := cfg.Accounts[:0]
	for _, a := range cfg.Accounts {
		if a.ID == id {
			found = true
			continue
		}
		next = append(next, a)
	}
	if !found {
		return fmt.Errorf("account %q not found", id)
	}

	cfg.Accounts = next
	if cfg.ActiveAccount != nil && *cfg.ActiveAccount == id {
		cfg.ActiveAccount = nil
	}
	if err := saveConfig(cfg); err != nil {
		return err
	}

	_ = os.Remove(paths.ProfilePath(id))
	return nil
}

func ProfileSaved(id paths.AccountID) bool {
	_, err := os.Stat(paths.ProfilePath(id))
	return err == nil
}

func Exists(id paths.AccountID) bool {
	return ProfileSaved(id)
}

func Load(id paths.AccountID) (*Profile, error) {
	data, err := os.ReadFile(paths.ProfilePath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, nil
	}
	return &p, nil
}

type SaveOptions struct {
	Label string
}

func SaveCurrentAs(id paths.AccountID, opts SaveOptions) (*Profile, error) {
	label := strings.TrimSpace(opts.Label)
	if !AccountExists(id) {
		if label == "" {
			return nil, fmt.Errorf("new account %q needs --label", id)
		}
		if err := AddAccount(id, label); err != nil {
			return nil, err
		}
	} else if label != "" {
		if err := RegisterAccount(id, label); err != nil {
			return nil, err
		}
	}

	db, err := database.Open(paths.CursorStateDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	authKeys, err := database.ReadAuthKeys(db)
	if err != nil {
		return nil, err
	}

	if authKeys["cursorAuth/accessToken"] == "" {
		return nil, fmt.Errorf("no Cursor auth session found — log into Cursor first, then save")
	}

	var email *string
	if e := authKeys["cursorAuth/cachedEmail"]; e != "" {
		email = &e
	}

	profile := &Profile{
		ID:       id,
		Label:    AccountLabel(id),
		Email:    email,
		SavedAt:  time.Now().UTC().Format(time.RFC3339),
		AuthKeys: authKeys,
	}

	if err := ensureDataDir(); err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths.ProfilePath(id), data, 0o600); err != nil {
		return nil, err
	}

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	cfg.ActiveAccount = &id
	if err := saveConfig(cfg); err != nil {
		return nil, err
	}

	return profile, nil
}

func ActiveAccount() *paths.AccountID {
	cfg, err := loadConfig()
	if err != nil {
		return nil
	}
	return cfg.ActiveAccount
}

func CurrentEmail() (string, error) {
	db, err := database.Open(paths.CursorStateDBPath())
	if err != nil {
		return "", err
	}
	defer db.Close()
	return database.ReadCurrentEmail(db)
}

func Restore(id paths.AccountID) (*Profile, error) {
	profile, err := Load(id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf(`no saved profile for "%s" — run: cursor-switch save %s`, AccountLabel(id), id)
	}
	if profile.AuthKeys["cursorAuth/accessToken"] == "" {
		return nil, fmt.Errorf(`profile "%s" has no access token`, AccountLabel(id))
	}

	db, err := database.Open(paths.CursorStateDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err := database.WriteAuthKeys(db, profile.AuthKeys); err != nil {
		return nil, err
	}

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	cfg.ActiveAccount = &id
	if err := saveConfig(cfg); err != nil {
		return nil, err
	}

	return profile, nil
}

func AutoSaveActive() {
	active := ActiveAccount()
	if active == nil {
		return
	}
	_, _ = SaveCurrentAs(*active, SaveOptions{})
}
