package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/reloadlife/cursor-account-switcher/internal/platform/keychain"
)

type ClaudeConfig struct {
	CredentialsPath func() string
	ConfigPath      func() string
	KeychainService string
}

type Claude struct {
	cfg ClaudeConfig
}

func NewClaude(cfg ClaudeConfig) *Claude {
	return &Claude{cfg: cfg}
}

type claudeCredentials struct {
	ClaudeAiOauth json.RawMessage `json:"claudeAiOauth"`
	McpOAuth      json.RawMessage `json:"mcpOAuth,omitempty"`
}

type claudeConfigFile struct {
	OAuthAccount json.RawMessage `json:"oauthAccount"`
}

func (c *Claude) Read() (Data, error) {
	oauth, _, err := c.readOAuth()
	if err != nil {
		return Data{}, err
	}

	credsJSON, err := json.Marshal(claudeCredentials{ClaudeAiOauth: oauth})
	if err != nil {
		return Data{}, err
	}

	files := map[string][]byte{
		c.cfg.CredentialsPath(): credsJSON,
	}

	if accountMeta, err := c.readOAuthAccount(); err == nil && len(accountMeta) > 0 {
		files[c.cfg.ConfigPath()] = accountMeta
	}

	data := Data{Files: files}
	return data, nil
}

func (c *Claude) Write(data Data) error {
	credsPath := c.cfg.CredentialsPath()
	credsRaw, ok := data.Files[credsPath]
	if !ok || len(credsRaw) == 0 {
		return fmt.Errorf("profile missing Claude credentials")
	}

	var creds claudeCredentials
	if err := json.Unmarshal(credsRaw, &creds); err != nil {
		return err
	}
	if len(creds.ClaudeAiOauth) == 0 {
		return fmt.Errorf("profile missing claudeAiOauth")
	}

	if err := os.MkdirAll(filepath.Dir(credsPath), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(credsPath, credsRaw, 0o600); err != nil {
		return err
	}

	if accountMeta, ok := data.Files[c.cfg.ConfigPath()]; ok && len(accountMeta) > 0 {
		if err := c.writeOAuthAccount(accountMeta); err != nil {
			return err
		}
	}

	if keychain.Supported() {
		merged, err := c.mergeKeychainOAuth(creds.ClaudeAiOauth)
		if err != nil {
			return err
		}
		return keychain.Store(keychain.Entry{
			Service: c.cfg.KeychainService,
			Account: os.Getenv("USER"),
			Secret:  merged,
		})
	}

	return nil
}

func (c *Claude) Clear() error {
	_ = os.Remove(c.cfg.CredentialsPath())
	// Strip oauthAccount from config if present; leave other settings intact.
	if path := c.cfg.ConfigPath(); path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg map[string]json.RawMessage
			if json.Unmarshal(data, &cfg) == nil {
				delete(cfg, "oauthAccount")
				if out, err := json.MarshalIndent(cfg, "", "  "); err == nil {
					_ = os.WriteFile(path, out, 0o600)
				}
			}
		}
	}
	if keychain.Supported() {
		_ = keychain.Delete(c.cfg.KeychainService, os.Getenv("USER"))
		_ = keychain.Delete(c.cfg.KeychainService, "")
	}
	return nil
}

func (c *Claude) Validate(data Data) error {
	credsRaw := data.Files[c.cfg.CredentialsPath()]
	if len(credsRaw) == 0 {
		return fmt.Errorf("no Claude Code auth session found — sign in first, then save")
	}
	var creds claudeCredentials
	if err := json.Unmarshal(credsRaw, &creds); err != nil {
		return fmt.Errorf("invalid Claude credentials in profile")
	}
	if len(creds.ClaudeAiOauth) == 0 {
		return fmt.Errorf("profile has no Claude OAuth token")
	}
	var oauth map[string]any
	if err := json.Unmarshal(creds.ClaudeAiOauth, &oauth); err != nil {
		return fmt.Errorf("invalid claudeAiOauth in profile")
	}
	if token, _ := oauth["accessToken"].(string); token == "" {
		return fmt.Errorf("profile has no Claude access token")
	}
	return nil
}

func (c *Claude) Identifier(data Data) string {
	if meta := data.Files[c.cfg.ConfigPath()]; len(meta) > 0 {
		if email := claudeEmailFromAccountMeta(meta); email != "" {
			return email
		}
	}
	if email, _ := c.readLiveEmail(); email != "" {
		return email
	}
	credsRaw := data.Files[c.cfg.CredentialsPath()]
	if len(credsRaw) > 0 {
		if sub := claudeSubscriptionFromCreds(credsRaw); sub != "" {
			return sub + " account"
		}
	}
	return ""
}

func (c *Claude) readOAuth() (json.RawMessage, string, error) {
	if keychain.Supported() {
		secret, err := keychain.Find(c.cfg.KeychainService, "")
		if err == nil && secret != "" {
			oauth, err := extractClaudeAiOauth([]byte(secret))
			if err != nil {
				return nil, "", err
			}
			return oauth, secret, nil
		}
	}

	path := c.cfg.CredentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("no Claude Code auth session found — run `claude` and sign in first")
		}
		return nil, "", err
	}

	oauth, err := extractClaudeAiOauth(data)
	if err != nil {
		return nil, "", err
	}
	return oauth, "", nil
}

func (c *Claude) readOAuthAccount() (json.RawMessage, error) {
	path := c.cfg.ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	meta, ok := cfg["oauthAccount"]
	if !ok || len(meta) == 0 {
		return nil, fmt.Errorf("oauthAccount not found")
	}
	return meta, nil
}

func (c *Claude) readLiveEmail() (string, error) {
	meta, err := c.readOAuthAccount()
	if err != nil {
		return "", err
	}
	return claudeEmailFromAccountMeta(meta), nil
}

func (c *Claude) writeOAuthAccount(accountMeta json.RawMessage) error {
	path := c.cfg.ConfigPath()
	var cfg map[string]json.RawMessage
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cfg = map[string]json.RawMessage{}
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return err
		}
	}

	cfg["oauthAccount"] = accountMeta
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}

func (c *Claude) mergeKeychainOAuth(oauth json.RawMessage) (string, error) {
	existing := map[string]json.RawMessage{}
	if keychain.Supported() {
		if secret, err := keychain.Find(c.cfg.KeychainService, ""); err == nil && secret != "" {
			_ = json.Unmarshal([]byte(secret), &existing)
		}
	}
	existing["claudeAiOauth"] = oauth
	out, err := json.Marshal(existing)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func extractClaudeAiOauth(raw []byte) (json.RawMessage, error) {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	oauth, ok := parsed["claudeAiOauth"]
	if !ok || len(oauth) == 0 {
		return nil, fmt.Errorf("no claudeAiOauth in Claude credentials — sign in with `claude` first")
	}
	return oauth, nil
}

func claudeEmailFromAccountMeta(meta json.RawMessage) string {
	var account map[string]any
	if err := json.Unmarshal(meta, &account); err != nil {
		return ""
	}
	if email, ok := account["emailAddress"].(string); ok {
		return email
	}
	return IdentifierFromJSON(meta)
}

func claudeSubscriptionFromCreds(raw []byte) string {
	var creds claudeCredentials
	if err := json.Unmarshal(raw, &creds); err != nil {
		return ""
	}
	var oauth map[string]any
	if err := json.Unmarshal(creds.ClaudeAiOauth, &oauth); err != nil {
		return ""
	}
	if sub, ok := oauth["subscriptionType"].(string); ok {
		return sub
	}
	return ""
}
