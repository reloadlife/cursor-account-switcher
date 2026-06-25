package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

type AccountID string

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}

func CursorStateDBPath() string {
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage", "state.vscdb")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Cursor", "User", "globalStorage", "state.vscdb")
	default:
		return filepath.Join(home, ".config", "Cursor", "User", "globalStorage", "state.vscdb")
	}
}

func CursorAppPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/Cursor.app"
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = homeDir()
		}
		return filepath.Join(localAppData, "Programs", "cursor", "Cursor.exe")
	default:
		return "cursor"
	}
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

func ProfilePath(id AccountID) string {
	return filepath.Join(DataDir(), "profiles", string(id)+".json")
}

func ConfigPath() string {
	return filepath.Join(DataDir(), "config.json")
}
