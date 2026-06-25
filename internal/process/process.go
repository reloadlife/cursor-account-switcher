package process

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/reloadlife/cursor-account-switcher/internal/paths"
)

func isCursorRunning() bool {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq Cursor.exe").Output()
		return err == nil && strings.Contains(string(out), "Cursor.exe")
	default:
		out, err := exec.Command("pgrep", "-f", "Cursor").Output()
		return err == nil && strings.TrimSpace(string(out)) != ""
	}
}

func ForceQuitCursor() error {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("taskkill", "/F", "/IM", "Cursor.exe").Run()
	case "darwin":
		names := []string{
			"Cursor",
			"Cursor Helper",
			"Cursor Helper (Renderer)",
			"Cursor Helper (GPU)",
			"Cursor Helper (Plugin)",
		}
		for _, name := range names {
			_ = exec.Command("killall", "-9", name).Run()
		}
	default:
		_ = exec.Command("pkill", "-9", "-f", "cursor").Run()
	}

	for i := 0; i < 30; i++ {
		if !isCursorRunning() {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("cursor did not exit. Close it manually and retry")
}

func StartCursor() error {
	appPath := paths.CursorAppPath()

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "-a", appPath).Start()
	case "windows":
		return exec.Command(appPath).Start()
	default:
		return exec.Command(appPath).Start()
	}
}
