package profiles

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/reloadlife/cursor-account-switcher/internal/paths"
	"github.com/reloadlife/cursor-account-switcher/internal/platform"
	"github.com/reloadlife/cursor-account-switcher/internal/platform/auth"
	"github.com/reloadlife/cursor-account-switcher/internal/platform/keychain"
)

type AccountDef struct {
	ID    paths.AccountID `json:"id"`
	Label string          `json:"label"`
}

type Profile struct {
	ID       paths.AccountID    `json:"id"`
	Platform platform.ID      `json:"platform,omitempty"`
	Label    string             `json:"label"`
	Email    *string            `json:"email"`
	SavedAt  string             `json:"savedAt"`
	AuthKeys map[string]string  `json:"authKeys,omitempty"`
	AuthFiles map[string][]byte `json:"authFiles,omitempty"`
	Keychain []keychain.Entry   `json:"keychain,omitempty"`
}

type Config struct {
	ActiveAccount *paths.AccountID `json:"activeAccount"`
	Accounts      []AccountDef     `json:"accounts"`
}

type GlobalConfig struct {
	ActivePlatform platform.ID `json:"activePlatform"`
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

func LoadGlobalConfig() (GlobalConfig, error) {
	data, err := os.ReadFile(paths.GlobalConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return GlobalConfig{ActivePlatform: platform.Cursor}, nil
		}
		return GlobalConfig{}, err
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return GlobalConfig{}, err
	}
	if cfg.ActivePlatform == "" {
		cfg.ActivePlatform = platform.Cursor
	}
	return cfg, nil
}

func SaveGlobalConfig(cfg GlobalConfig) error {
	if err := os.MkdirAll(paths.DataDir(), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.GlobalConfigPath(), data, 0o600)
}

func InitPlatform() error {
	if err := migrateLegacyLayout(); err != nil {
		return err
	}
	global, err := LoadGlobalConfig()
	if err != nil {
		return err
	}
	return platform.SetCurrent(global.ActivePlatform)
}

func migrateLegacyLayout() error {
	legacyProfiles := paths.LegacyProfilesDir()
	cursorProfiles := filepath.Join(paths.PlatformDataDir(platform.Cursor), "profiles")
	legacyConfig := paths.LegacyConfigPath()
	cursorConfig := paths.ConfigPathFor(platform.Cursor)

	if _, err := os.Stat(legacyProfiles); err == nil {
		if _, err := os.Stat(cursorProfiles); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(cursorProfiles), 0o700); err != nil {
				return err
			}
			if err := os.Rename(legacyProfiles, cursorProfiles); err != nil {
				if err := copyDir(legacyProfiles, cursorProfiles); err != nil {
					return err
				}
			}
		}
	}

	data, err := os.ReadFile(legacyConfig)
	if err != nil {
		return nil
	}

	var old Config
	if err := json.Unmarshal(data, &old); err != nil {
		return nil
	}
	if len(old.Accounts) == 0 {
		return nil
	}

	if _, err := os.Stat(cursorConfig); os.IsNotExist(err) {
		if err := ensureDataDir(); err != nil {
			return err
		}
		if err := saveConfig(old); err != nil {
			return err
		}
	}

	global := GlobalConfig{ActivePlatform: platform.Cursor}
	return SaveGlobalConfig(global)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, infoMode(src))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func infoMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0o600
	}
	return info.Mode().Perm()
}

func ensureDataDir() error {
	if err := os.MkdirAll(paths.ProfilesDir(), 0o700); err != nil {
		return err
	}
	return os.MkdirAll(paths.PlatformDataDir(platform.Current()), 0o700)
}

func loadConfig() (Config, error) {
	if err := ensureDataDir(); err != nil {
		return Config{}, err
	}

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

func profileFromAuth(id paths.AccountID, authData auth.Data) *Profile {
	var email *string
	if idStr := currentPlatform().Auth.Identifier(authData); idStr != "" {
		email = &idStr
	}

	return &Profile{
		ID:        id,
		Platform:  platform.Current(),
		Label:     AccountLabel(id),
		Email:     email,
		SavedAt:   time.Now().UTC().Format(time.RFC3339),
		AuthKeys:  authData.Keys,
		AuthFiles: authData.Files,
		Keychain:  authData.Keychain,
	}
}

func authFromProfile(profile *Profile) auth.Data {
	return auth.Data{
		Keys:     profile.AuthKeys,
		Files:    profile.AuthFiles,
		Keychain: profile.Keychain,
	}
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

	p := currentPlatform()
	authData, err := p.Auth.Read()
	if err != nil {
		return nil, err
	}
	if err := p.Auth.Validate(authData); err != nil {
		return nil, err
	}

	profile := profileFromAuth(id, authData)

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

func CurrentIdentifier() (string, error) {
	p := currentPlatform()
	authData, err := p.Auth.Read()
	if err != nil {
		return "", err
	}
	return p.Auth.Identifier(authData), nil
}

func CurrentEmail() (string, error) {
	return CurrentIdentifier()
}

func Restore(id paths.AccountID) (*Profile, error) {
	profile, err := Load(id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf(`no saved profile for "%s" — run: cursor-switch save %s`, AccountLabel(id), id)
	}

	authData := authFromProfile(profile)
	p := currentPlatform()
	if err := p.Auth.Validate(authData); err != nil {
		return nil, fmt.Errorf(`profile "%s" has invalid credentials: %w`, AccountLabel(id), err)
	}

	if err := p.Auth.Write(authData); err != nil {
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

func currentPlatform() *platform.Platform {
	p, err := platform.CurrentPlatform()
	if err != nil {
		return platform.All()[platform.Cursor]
	}
	return p
}

func SetActivePlatform(id platform.ID) error {
	if err := platform.SetCurrent(id); err != nil {
		return err
	}
	return SaveGlobalConfig(GlobalConfig{ActivePlatform: id})
}
