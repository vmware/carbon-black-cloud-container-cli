// Package user manages the user selection commands.
package user

import (
	"github.com/spf13/cobra"
)

// Cmd will return the user command.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage cbctl user profiles",
		Long:  `Manage cbctl Black user profiles.`,
	}

	cmd.AddCommand(ListCmd())
	cmd.AddCommand(AddCmd())
	cmd.AddCommand(DelCmd())

	return cmd
}
