package platform

import (
	"regexp"
	"runtime"
)

// cursorLinuxPgrepPatterns match Cursor IDE processes on Linux without hitting
// unrelated commands such as cursor-switch or generic "cursor" substrings in
// shell one-liners. Patterns are used with pgrep/pkill -f (ERE).
func cursorLinuxPgrepPatterns() []string {
	return []string{
		"cursor.mjs",
		"/share/cursor/resources/",
		".config/Cursor",
	}
}

func cursorProcessConfig() *ProcessConfig {
	cfg := &ProcessConfig{
		DisplayName: "Cursor",
		AppPath:     cursorAppPath,
		WindowsExe:  "Cursor.exe",
	}
	switch runtime.GOOS {
	case "darwin":
		cfg.KillNames = []string{
			"Cursor",
			"Cursor Helper",
			"Cursor Helper (Renderer)",
			"Cursor Helper (GPU)",
			"Cursor Helper (Plugin)",
		}
		cfg.PgrepMatch = "Cursor"
	default:
		if runtime.GOOS == "linux" {
			cfg.KillNames = []string{"cursor"}
			cfg.PgrepMatches = cursorLinuxPgrepPatterns()
		} else {
			// Other Unix (e.g. FreeBSD): pgrep full command line, capitalized mac-style names.
			cfg.KillNames = []string{"Cursor", "cursor"}
			cfg.PgrepMatches = append(cursorLinuxPgrepPatterns(), "Cursor")
		}
	}
	return cfg
}

// pgrepPatternMatches reports whether cmdline would match pgrep -f pattern.
// Used by tests; pgrep uses POSIX ERE where unescaped . matches any character.
func pgrepPatternMatches(cmdline, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(cmdline)
}

func matchesAnyPgrepPattern(cmdline string, patterns []string) bool {
	for _, pattern := range patterns {
		if pgrepPatternMatches(cmdline, pattern) {
			return true
		}
	}
	return false
}
