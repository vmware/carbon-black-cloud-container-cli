// Package auth manages the credentials subcommand and utility functions.
package auth

import (
	"github.com/spf13/cobra"
)

// Cmd return the command related to credential.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Set auth for cbctl",
		Long:  `Set auth for cbctl`,
	}

	cmd.AddCommand(SetCbAPIAccessCmd())

	return cmd
}
