package auth

import (
	"encoding/json"
	"strings"

	"github.com/reloadlife/cursor-account-switcher/internal/platform/keychain"
)

type Data struct {
	Keys     map[string]string    `json:"keys,omitempty"`
	Files    map[string][]byte      `json:"files,omitempty"`
	Keychain []keychain.Entry       `json:"keychain,omitempty"`
}

type Backend interface {
	Read() (Data, error)
	Write(Data) error
	Validate(Data) error
	Identifier(Data) string
}

func IdentifierFromJSON(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return ""
	}
	for _, path := range [][]string{
		{"email"},
		{"account", "email"},
		{"account", "login"},
		{"login"},
		{"user", "email"},
		{"oauthAccount", "emailAddress"},
		{"claudeAiOauth", "subscriptionType"},
		{"tokens", "account_id"},
	} {
		if v := nestedString(data, path); v != "" {
			return v
		}
	}
	return ""
}

func nestedString(data map[string]any, path []string) string {
	var current any = data
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = m[key]
		if !ok {
			return ""
		}
	}
	switch v := current.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return ""
	}
}
