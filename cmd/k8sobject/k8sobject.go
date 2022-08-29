// Package k8sobject manages the k8s-resource analysis subcommands.
package k8sobject

import (
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/presenter"
)

type (
	presenterOption = presenter.Option
)

var opts struct {
	presenterOption
}

// Cmd return the command related to k8s-resource analysis.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s-object",
		Short: "Commands related to k8s-object analysis",
		Long:  `Commands related to k8s-object analysis`,
	}

	cmd.AddCommand(ValidateCmd())

	cmd.PersistentFlags().StringVarP(
		&opts.OutputFormat, "output", "o", "table", "output format of the result")

	return cmd
}
