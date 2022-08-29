package cmd

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.bit9.local/octarine/cbctl/internal"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/config"
	"gitlab.bit9.local/octarine/cbctl/internal/util/memorytool"
	"gitlab.bit9.local/octarine/cbctl/internal/version"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
)

var ctx, cancel = context.WithCancel(context.Background())

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   internal.ApplicationName,
	Short: "Carbon Black's instrumentation client",
	Long:  `A client CLI for image scanning, and instrumenting Carbon Black services.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		go memorytool.ReadMemoryStats(ctx)
		checkNewVersion()
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		cancel()
		return config.PersistConfig()
	},
}

// checkNewVersion will check if a new version is available.
func checkNewVersion() {
	isAvailable, newVersion, err := version.IsUpdateAvailable()
	if err != nil {
		errMsg := "Failed to detect new version"
		e := cberr.NewError(cberr.HTTPConnectionErr, errMsg, err)
		logrus.Errorln(e)

		return
	}

	if isAvailable {
		bus.Publish(bus.NewVersionEvent(newVersion))
	}
}
