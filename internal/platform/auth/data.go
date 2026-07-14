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
	// Clear removes live auth so a fresh sign-in can fill an empty profile slot.
	Clear() error
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
	// Grok OIDC and similar: top-level map of provider → { email, … }
	if v := firstEmailDeep(data, 0); v != "" {
		return v
	}
	return ""
}

func firstEmailDeep(v any, depth int) string {
	if depth > 6 {
		return ""
	}
	switch t := v.(type) {
	case map[string]any:
		for _, key := range []string{"email", "emailAddress", "login"} {
			if s, ok := t[key].(string); ok && strings.Contains(s, "@") {
				return strings.TrimSpace(s)
			}
		}
		for _, child := range t {
			if s := firstEmailDeep(child, depth+1); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range t {
			if s := firstEmailDeep(child, depth+1); s != "" {
				return s
			}
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
