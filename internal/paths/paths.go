package paths

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/reloadlife/cursor-account-switcher/internal/platform"
)

type AccountID string

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}

func DataDir() string {
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, ".cursor-account-switcher")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "cursor-account-switcher")
	default:
		return filepath.Join(home, ".config", "cursor-account-switcher")
	}
}

func GlobalConfigPath() string {
	return filepath.Join(DataDir(), "config.json")
}

func PlatformDataDir(id platform.ID) string {
	return filepath.Join(DataDir(), string(id))
}

func ConfigPath() string {
	return ConfigPathFor(platform.Current())
}

func ConfigPathFor(id platform.ID) string {
	return filepath.Join(PlatformDataDir(id), "config.json")
}

func ProfilePath(id AccountID) string {
	return ProfilePathFor(platform.Current(), id)
}

func ProfilePathFor(platformID platform.ID, id AccountID) string {
	return filepath.Join(PlatformDataDir(platformID), "profiles", string(id)+".json")
}

func ProfilesDir() string {
	return filepath.Join(PlatformDataDir(platform.Current()), "profiles")
}

func LegacyConfigPath() string {
	return filepath.Join(DataDir(), "config.json")
}

func LegacyProfilesDir() string {
	return filepath.Join(DataDir(), "profiles")
}
