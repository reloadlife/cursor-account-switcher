package keychain

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Entry struct {
	Service string `json:"service"`
	Account string `json:"account"`
	Secret  string `json:"secret"`
}

func Supported() bool {
	switch runtime.GOOS {
	case "darwin":
		return true
	case "linux":
		_, err := exec.LookPath("secret-tool")
		return err == nil
	default:
		return false
	}
}

func Find(service, account string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		args := []string{"find-generic-password", "-s", service, "-w"}
		if account != "" {
			args = append(args, "-a", account)
		}
		out, err := exec.Command("security", args...).Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	case "linux":
		args := []string{"lookup", "service", service}
		if account != "" {
			args = append(args, "account", account)
		}
		out, err := exec.Command("secret-tool", args...).Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	default:
		return "", fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

func Store(entry Entry) error {
	if entry.Service == "" {
		return fmt.Errorf("keychain service is required")
	}
	_ = Delete(entry.Service, entry.Account)

	switch runtime.GOOS {
	case "darwin":
		account := entry.Account
		if account == "" {
			account = os.Getenv("USER")
		}
		return exec.Command(
			"security", "add-generic-password",
			"-s", entry.Service,
			"-a", account,
			"-w", entry.Secret,
		).Run()
	case "linux":
		args := []string{"store", "--label", entry.Service, "service", entry.Service}
		if entry.Account != "" {
			args = append(args, "account", entry.Account)
		}
		cmd := exec.Command("secret-tool", args...)
		cmd.Stdin = strings.NewReader(entry.Secret)
		return cmd.Run()
	default:
		return fmt.Errorf("keychain not supported on %s", runtime.GOOS)
	}
}

func Delete(service, account string) error {
	switch runtime.GOOS {
	case "darwin":
		args := []string{"delete-generic-password", "-s", service}
		if account != "" {
			args = append(args, "-a", account)
		}
		_ = exec.Command("security", args...).Run()
		return nil
	case "linux":
		args := []string{"clear", "service", service}
		if account != "" {
			args = append(args, "account", account)
		}
		_ = exec.Command("secret-tool", args...).Run()
		return nil
	default:
		return nil
	}
}

func Restore(entries []Entry) error {
	for _, entry := range entries {
		if entry.Secret == "" {
			continue
		}
		if err := Store(entry); err != nil {
			return err
		}
	}
	return nil
}
