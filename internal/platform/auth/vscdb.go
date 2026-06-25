package auth

import (
	"fmt"
	"os"

	"github.com/reloadlife/cursor-account-switcher/internal/database"
)

type VSCDBConfig struct {
	DBPath       func() string
	KeyPrefix    string
	KeyPatterns  []string
	RequiredKey  string
	EmailKey     string
	PlatformName string
}

type VSCDB struct {
	cfg VSCDBConfig
}

func NewVSCDB(cfg VSCDBConfig) *VSCDB {
	return &VSCDB{cfg: cfg}
}

func (v *VSCDB) Read() (Data, error) {
	path := v.cfg.DBPath()
	db, err := database.Open(path)
	if err != nil {
		return Data{}, fmt.Errorf("%s state database not found: %s", v.cfg.PlatformName, path)
	}
	defer db.Close()

	keys, err := database.ReadKeys(db, v.cfg.KeyPrefix, v.cfg.KeyPatterns)
	if err != nil {
		return Data{}, err
	}

	return Data{Keys: keys}, nil
}

func (v *VSCDB) Write(data Data) error {
	path := v.cfg.DBPath()
	db, err := database.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := database.WriteKeys(db, v.cfg.KeyPrefix, v.cfg.KeyPatterns, data.Keys); err != nil {
		return err
	}
	return nil
}

func (v *VSCDB) Validate(data Data) error {
	if v.cfg.RequiredKey != "" {
		if data.Keys[v.cfg.RequiredKey] == "" {
			return fmt.Errorf("no %s auth session found — sign in first, then save", v.cfg.PlatformName)
		}
		return nil
	}
	if len(data.Keys) > 0 {
		return nil
	}
	return fmt.Errorf("no %s auth session found — sign in first, then save", v.cfg.PlatformName)
}

func (v *VSCDB) Identifier(data Data) string {
	if v.cfg.EmailKey != "" {
		if email := data.Keys[v.cfg.EmailKey]; email != "" {
			return email
		}
	}
	for _, value := range data.Keys {
		if id := IdentifierFromJSON([]byte(value)); id != "" {
			return id
		}
	}
	return ""
}

type FileConfig struct {
	Paths        []func() string
	PlatformName string
}

type File struct {
	cfg FileConfig
}

func NewFile(cfg FileConfig) *File {
	return &File{cfg: cfg}
}

func (f *File) Read() (Data, error) {
	files := make(map[string][]byte)
	for _, pathFn := range f.cfg.Paths {
		path := pathFn()
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Data{}, err
		}
		if len(data) > 0 {
			files[path] = data
		}
	}
	if len(files) == 0 {
		return Data{}, fmt.Errorf(
			"no %s auth session found — sign in first, then save",
			f.cfg.PlatformName,
		)
	}
	return Data{Files: files}, nil
}

func (f *File) Write(data Data) error {
	for path, content := range data.Files {
		if err := os.MkdirAll(dirOf(path), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) Validate(data Data) error {
	if len(data.Files) == 0 {
		return fmt.Errorf("no %s auth session found — sign in first, then save", f.cfg.PlatformName)
	}
	for _, content := range data.Files {
		if len(content) > 0 {
			return nil
		}
	}
	return fmt.Errorf("profile has no %s credentials", f.cfg.PlatformName)
}

func (f *File) Identifier(data Data) string {
	for _, content := range data.Files {
		if id := IdentifierFromJSON(content); id != "" {
			return id
		}
	}
	return ""
}

func dirOf(path string) string {
	i := len(path) - 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	if i <= 0 {
		return "."
	}
	return path[:i]
}
