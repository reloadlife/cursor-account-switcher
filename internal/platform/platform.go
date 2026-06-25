package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/reloadlife/cursor-account-switcher/internal/platform/auth"
	"github.com/reloadlife/cursor-account-switcher/internal/platform/keychain"
)

type Platform struct {
	ID          ID
	Name        string
	Description string
	Auth        auth.Backend
	Process     *ProcessConfig
}

type ProcessConfig struct {
	DisplayName string
	AppPath     func() string
	KillNames   []string
	PgrepMatch  string
	WindowsExe  string
}

func All() map[ID]*Platform {
	return map[ID]*Platform{
		Cursor: cursorPlatform(),
		Claude: claudePlatform(),
		Codex:  codexPlatform(),
		VSCode: vscodePlatform(),
	}
}

func List() []*Platform {
	all := All()
	order := []ID{Cursor, Claude, Codex, VSCode}
	out := make([]*Platform, 0, len(order))
	for _, id := range order {
		if p, ok := all[id]; ok {
			out = append(out, p)
		}
	}
	return out
}

func (p *Platform) NeedsRestart() bool {
	return p.Process != nil
}

func (p *Platform) ForceQuit() error {
	if p.Process == nil {
		return nil
	}
	cfg := p.Process

	switch runtime.GOOS {
	case "windows":
		if cfg.WindowsExe != "" {
			_ = exec.Command("taskkill", "/F", "/IM", cfg.WindowsExe).Run()
		}
	default:
		for _, name := range cfg.KillNames {
			_ = exec.Command("killall", "-9", name).Run()
		}
		if cfg.PgrepMatch != "" {
			_ = exec.Command("pkill", "-9", "-f", cfg.PgrepMatch).Run()
		}
	}

	for i := 0; i < 30; i++ {
		if !p.IsRunning() {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmtError("%s did not exit — close it manually and retry", cfg.DisplayName)
}

func (p *Platform) Start() error {
	if p.Process == nil {
		return nil
	}
	appPath := p.Process.AppPath()

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "-a", appPath).Start()
	case "windows":
		return exec.Command(appPath).Start()
	default:
		return exec.Command(appPath).Start()
	}
}

func (p *Platform) IsRunning() bool {
	if p.Process == nil {
		return false
	}
	cfg := p.Process

	switch runtime.GOOS {
	case "windows":
		if cfg.WindowsExe == "" {
			return false
		}
		out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+cfg.WindowsExe).Output()
		return err == nil && strings.Contains(string(out), cfg.WindowsExe)
	default:
		if cfg.PgrepMatch != "" {
			out, err := exec.Command("pgrep", "-f", cfg.PgrepMatch).Output()
			return err == nil && strings.TrimSpace(string(out)) != ""
		}
		for _, name := range cfg.KillNames {
			out, err := exec.Command("pgrep", "-x", name).Output()
			if err == nil && strings.TrimSpace(string(out)) != "" {
				return true
			}
		}
	}
	return false
}

func cursorPlatform() *Platform {
	return &Platform{
		ID:          Cursor,
		Name:        "Cursor",
		Description: "Cursor IDE auth (state.vscdb)",
		Auth: auth.NewVSCDB(auth.VSCDBConfig{
			DBPath:       cursorStateDBPath,
			KeyPrefix:    "cursorAuth/",
			RequiredKey:  "cursorAuth/accessToken",
			EmailKey:     "cursorAuth/cachedEmail",
			PlatformName: "Cursor",
		}),
		Process: &ProcessConfig{
			DisplayName: "Cursor",
			AppPath:     cursorAppPath,
			KillNames: []string{
				"Cursor",
				"Cursor Helper",
				"Cursor Helper (Renderer)",
				"Cursor Helper (GPU)",
				"Cursor Helper (Plugin)",
			},
			PgrepMatch: "Cursor",
			WindowsExe: "Cursor.exe",
		},
	}
}

func claudePlatform() *Platform {
	return &Platform{
		ID:          Claude,
		Name:        "Claude Code",
		Description: "Claude Code CLI auth (credentials + ~/.claude.json account)",
		Auth: auth.NewClaude(auth.ClaudeConfig{
			CredentialsPath: claudeCredentialsPath,
			ConfigPath:      claudeConfigPath,
			KeychainService: "Claude Code-credentials",
		}),
	}
}

func claudeConfigPath() string {
	return filepath.Join(homeDir(), ".claude.json")
}

func codexPlatform() *Platform {
	return &Platform{
		ID:          Codex,
		Name:        "Codex",
		Description: "OpenAI Codex CLI auth (~/.codex/auth.json)",
		Auth: auth.NewFile(auth.FileConfig{
			Paths:        []func() string{codexAuthPath},
			PlatformName: "Codex",
		}),
	}
}

func vscodePlatform() *Platform {
	vscdb := auth.NewVSCDB(auth.VSCDBConfig{
		DBPath:      vscodeStateDBPath,
		KeyPatterns: []string{"%github%", "secret://%GitHub%", "secret://%github%"},
		RequiredKey: "",
		PlatformName: "VS Code",
	})
	specs := []auth.KeychainSpec{}
	if keychain.Supported() {
		specs = append(specs, auth.KeychainSpec{
			Service: "vscodevscode.github-authentication",
			Account: "github.auth",
		})
	}
	return &Platform{
		ID:          VSCode,
		Name:        "VS Code",
		Description: "VS Code / GitHub Copilot auth (state.vscdb + keychain)",
		Auth: auth.NewNamedComposite("VS Code", vscdb, specs),
		Process: &ProcessConfig{
			DisplayName: "Visual Studio Code",
			AppPath:     vscodeAppPath,
			KillNames:   []string{"Code"},
			PgrepMatch:  "Visual Studio Code",
			WindowsExe:  "Code.exe",
		},
	}
}

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}

func cursorStateDBPath() string {
	return filepath.Join(appSupportDir("Cursor"), "User", "globalStorage", "state.vscdb")
}

func vscodeStateDBPath() string {
	return filepath.Join(appSupportDir("Code"), "User", "globalStorage", "state.vscdb")
}

func appSupportDir(app string) string {
	home := homeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", app)
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, app)
	default:
		return filepath.Join(home, ".config", app)
	}
}

func cursorAppPath() string {
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

func vscodeAppPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/Visual Studio Code.app"
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = homeDir()
		}
		return filepath.Join(localAppData, "Programs", "Microsoft VS Code", "Code.exe")
	default:
		return "code"
	}
}

func claudeCredentialsPath() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, ".credentials.json")
	}
	return filepath.Join(homeDir(), ".claude", ".credentials.json")
}

func codexAuthPath() string {
	if dir := os.Getenv("CODEX_HOME"); dir != "" {
		return filepath.Join(dir, "auth.json")
	}
	return filepath.Join(homeDir(), ".codex", "auth.json")
}
