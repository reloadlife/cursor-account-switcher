//go:build windows

package platform

import (
	"os/exec"
	"strings"
)

func unixForceQuit(cfg *ProcessConfig) {}

func unixIsRunning(cfg *ProcessConfig) bool { return false }

func windowsForceQuit(cfg *ProcessConfig) {
	if cfg.WindowsExe != "" {
		_ = exec.Command("taskkill", "/F", "/IM", cfg.WindowsExe).Run()
	}
}

func windowsIsRunning(cfg *ProcessConfig) bool {
	if cfg.WindowsExe == "" {
		return false
	}
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+cfg.WindowsExe).Output()
	return err == nil && strings.Contains(string(out), cfg.WindowsExe)
}
