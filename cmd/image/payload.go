package image

import (
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/version"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/presenter"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/scan"
)

// PayloadCmd will return the image scan request payload.
func PayloadCmd() *cobra.Command {
	payloadCmd := &cobra.Command{
		Use:  "payload <source>",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			go printPayload(args[0])
			terminalui.NewDisplay().DisplayEvents()
		},
	}

	payloadCmd.Hidden = true
	return payloadCmd
}

// printPayload will print the scan payload.
func printPayload(input string) {
	scanner := scan.NewScanner()
	generatedBom, imgLayers, err := scanner.ExtractDataFromImage(input, opts.scanOption)
	if err {
		return
	}

	versionInfo := version.GetCurrentVersion()
	payload := scan.NewAnalysisPayload(&generatedBom.Packages, imgLayers, "", "", false, versionInfo.SyftVersion, versionInfo.Version)

	opts.presenterOption.Limit = len(imgLayers)
	bus.Publish(bus.NewEvent(bus.PrintPayload, presenter.NewPresenter(&payload, opts.presenterOption), true))
}
