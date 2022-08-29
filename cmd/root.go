package cmd

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/memorytool"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/version"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
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
