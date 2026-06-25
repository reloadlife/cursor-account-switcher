package cmd

import (
	"fmt"

	"github.com/reloadlife/cursor-account-switcher/internal/app"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show saved profiles and current session",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(app.StatusText())
		return nil
	},
}
