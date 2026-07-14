package platform

type ID string

const (
	Cursor ID = "cursor"
	Claude ID = "claude"
	Codex  ID = "codex"
	Grok   ID = "grok"
	VSCode ID = "vscode"
)

var current = Cursor

func SetCurrent(id ID) error {
	if _, ok := All()[id]; !ok {
		return unknownPlatform(id)
	}
	current = id
	return nil
}

func Current() ID {
	return current
}

func Get(id ID) (*Platform, error) {
	p, ok := All()[id]
	if !ok {
		return nil, unknownPlatform(id)
	}
	return p, nil
}

func CurrentPlatform() (*Platform, error) {
	return Get(current)
}

func unknownPlatform(id ID) error {
	return fmtError("unknown platform %q — run: cursor-switch platform list", id)
}
