// Package image manages the image analysis subcommands.
package image

import (
	"github.com/spf13/cobra"
	"gitlab.bit9.local/octarine/cbctl/pkg/presenter"
	"gitlab.bit9.local/octarine/cbctl/pkg/scan"
)

type (
	scanOption      = scan.Option
	presenterOption = presenter.Option
)

var opts struct {
	scanOption
	presenterOption
}

const (
	fullTable      = 0
	defaultTimeout = 600
)

// Cmd return the command related to image analysis.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Commands related to image analysis",
		Long:  `Commands related to image analysis`,
	}

	cmd.AddCommand(ScanCmd())
	cmd.AddCommand(ValidateCmd())
	cmd.AddCommand(PackagesCmd())

	cmd.PersistentFlags().StringVarP(
		&opts.OutputFormat, "output", "o", "table", "output format of the result")
	cmd.PersistentFlags().BoolVar(
		&opts.ShouldCleanup, "cleanup", false, "clean up image (for docker only) after scanning")
	cmd.PersistentFlags().BoolVar(
		&opts.BypassDockerDaemon, "bypass-docker-daemon", false,
		"try to pull image without docker daemon")
	cmd.PersistentFlags().BoolVar(
		&opts.UseDockerDaemon, "use-docker", false, "deprecated - docker daemon is now the default")
	cmd.PersistentFlags().StringVar(
		&opts.Credential, "cred", "", "use `USERNAME[:PASSWORD]` for accessing the registry")
	cmd.PersistentFlags().IntVar(
		&opts.Timeout, "timeout", defaultTimeout, "set the duration (second) for the scan process")

	return cmd
}
