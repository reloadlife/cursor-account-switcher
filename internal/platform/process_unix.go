//go:build !windows

package platform

import (
	"os/exec"
	"strings"
)

func (cfg *ProcessConfig) pgrepPatterns() []string {
	if len(cfg.PgrepMatches) > 0 {
		return cfg.PgrepMatches
	}
	if cfg.PgrepMatch != "" {
		return []string{cfg.PgrepMatch}
	}
	return nil
}

func unixForceQuit(cfg *ProcessConfig) {
	for _, name := range cfg.KillNames {
		_ = exec.Command("killall", "-9", name).Run()
	}
	for _, pattern := range cfg.pgrepPatterns() {
		_ = exec.Command("pkill", "-9", "-f", pattern).Run()
	}
}

func unixIsRunning(cfg *ProcessConfig) bool {
	for _, pattern := range cfg.pgrepPatterns() {
		out, err := exec.Command("pgrep", "-f", pattern).Output()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return true
		}
	}
	for _, name := range cfg.KillNames {
		out, err := exec.Command("pgrep", "-x", name).Output()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return true
		}
	}
	return false
}

func windowsForceQuit(cfg *ProcessConfig) {}

func windowsIsRunning(cfg *ProcessConfig) bool { return false }
