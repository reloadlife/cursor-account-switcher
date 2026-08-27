package platform

import (
	"runtime"
	"testing"
)

func TestCursorLinuxPgrepPatterns(t *testing.T) {
	patterns := cursorLinuxPgrepPatterns()

	positive := []string{
		"/usr/lib/electron40/electron /usr/share/cursor/resources/app/cursor.mjs",
		"/usr/lib/electron40/electron --dns-result-order=ipv4first /usr/share/cursor/resources/app/extensions/cursor-always-local/dist/gitWorker.js",
		"/proc/self/exe --type=utility --user-data-dir=/home/alice/.config/Cursor --standard-schemes=vscode-webview",
		"/usr/share/cursor/resources/app/resources/helpers/cursorsandbox --run-proxy",
	}

	for _, cmdline := range positive {
		if !matchesAnyPgrepPattern(cmdline, patterns) {
			t.Errorf("expected Cursor IDE cmdline to match: %q", cmdline)
		}
	}

	negative := []string{
		"/home/alice/go/bin/cursor-switch switch work",
		"/home/alice/projects/cursor-account-switcher/cursor-switch status",
		"bash -c pgrep -af cursor",
		"/usr/bin/python3 /opt/some-app/manage_cursors.py",
		"/usr/bin/code /home/alice/project",
	}

	for _, cmdline := range negative {
		if matchesAnyPgrepPattern(cmdline, patterns) {
			t.Errorf("expected unrelated cmdline not to match: %q", cmdline)
		}
	}
}

func TestCursorProcessConfigByGOOS(t *testing.T) {
	cfg := cursorProcessConfig()
	if cfg.DisplayName != "Cursor" {
		t.Fatalf("DisplayName = %q, want Cursor", cfg.DisplayName)
	}
	if cfg.WindowsExe != "Cursor.exe" {
		t.Fatalf("WindowsExe = %q, want Cursor.exe", cfg.WindowsExe)
	}

	switch runtime.GOOS {
	case "darwin":
		if cfg.PgrepMatch != "Cursor" {
			t.Fatalf("darwin PgrepMatch = %q, want Cursor", cfg.PgrepMatch)
		}
		if len(cfg.KillNames) == 0 || cfg.KillNames[0] != "Cursor" {
			t.Fatalf("darwin KillNames = %v, want Cursor first", cfg.KillNames)
		}
	case "linux":
		if len(cfg.PgrepMatches) == 0 {
			t.Fatal("linux PgrepMatches is empty")
		}
		if len(cfg.KillNames) != 1 || cfg.KillNames[0] != "cursor" {
			t.Fatalf("linux KillNames = %v, want [cursor]", cfg.KillNames)
		}
	default:
		if len(cfg.PgrepMatches) == 0 {
			t.Fatal("unix PgrepMatches is empty")
		}
	}
}
