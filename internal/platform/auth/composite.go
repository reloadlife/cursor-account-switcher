package auth

import (
	"fmt"
	"os"

	"github.com/reloadlife/cursor-account-switcher/internal/platform/keychain"
)

type KeychainSpec struct {
	Service string
	Account string
}

type Composite struct {
	primary      Backend
	keychain     []KeychainSpec
	platformName string
	filePaths    []string
}

func NewComposite(primary Backend, specs []KeychainSpec) *Composite {
	return &Composite{primary: primary, keychain: specs}
}

func NewNamedComposite(name string, primary Backend, specs []KeychainSpec, filePaths ...string) *Composite {
	return &Composite{primary: primary, keychain: specs, platformName: name, filePaths: filePaths}
}

func (c *Composite) Read() (Data, error) {
	data, _ := c.primary.Read()

	for _, spec := range c.keychain {
		if !keychain.Supported() {
			continue
		}
		secret, err := keychain.Find(spec.Service, spec.Account)
		if err != nil || secret == "" {
			continue
		}
		account := spec.Account
		if account == "" {
			account = os.Getenv("USER")
		}
		data.Keychain = append(data.Keychain, keychain.Entry{
			Service: spec.Service,
			Account: account,
			Secret:  secret,
		})
	}

	if len(data.Files) == 0 && len(data.Keychain) > 0 && len(c.filePaths) > 0 {
		secret := data.Keychain[0].Secret
		if secret != "" {
			if data.Files == nil {
				data.Files = make(map[string][]byte)
			}
			data.Files[c.filePaths[0]] = []byte(secret)
		}
	}

	if len(data.Files) == 0 && len(data.Keys) == 0 && len(data.Keychain) == 0 {
		if c.platformName != "" {
			return Data{}, fmt.Errorf(
				"no %s auth session found — sign in first, then save",
				c.platformName,
			)
		}
		return Data{}, fmt.Errorf("no auth session found — sign in first, then save")
	}

	return data, nil
}

func (c *Composite) Write(data Data) error {
	if err := c.primary.Write(data); err != nil {
		return err
	}
	if len(data.Keychain) > 0 && keychain.Supported() {
		return keychain.Restore(data.Keychain)
	}
	return nil
}

func (c *Composite) Clear() error {
	if err := c.primary.Clear(); err != nil {
		return err
	}
	if keychain.Supported() {
		for _, spec := range c.keychain {
			account := spec.Account
			if account == "" {
				account = os.Getenv("USER")
			}
			_ = keychain.Delete(spec.Service, account)
			// also try empty account variant
			_ = keychain.Delete(spec.Service, "")
		}
	}
	return nil
}

func (c *Composite) Validate(data Data) error {
	if len(data.Keychain) > 0 {
		return nil
	}
	return c.primary.Validate(data)
}

func (c *Composite) Identifier(data Data) string {
	if id := c.primary.Identifier(data); id != "" {
		return id
	}
	for _, entry := range data.Keychain {
		if id := IdentifierFromJSON([]byte(entry.Secret)); id != "" {
			return id
		}
	}
	return ""
}
