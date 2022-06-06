package image

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.bit9.local/octarine/cbctl/internal"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/config"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui"
	"gitlab.bit9.local/octarine/cbctl/internal/util/printtool"
	"gitlab.bit9.local/octarine/cbctl/internal/util/tabletool"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
	"gitlab.bit9.local/octarine/cbctl/pkg/model/image"
	"gitlab.bit9.local/octarine/cbctl/pkg/presenter"
	"gitlab.bit9.local/octarine/cbctl/pkg/scan"
	"gitlab.bit9.local/octarine/cbctl/pkg/validate"
)

var (
	buildStep string
	namespace string

	validateScanHandler  *scan.Handler
	validateImageHandler *validate.ImageHandler
)

// ValidateCmd will return the image validate command.
func ValidateCmd() *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate <source>",
		Short: "Validate scanned image and generate violations report",
		Long: printtool.Tprintf(`Validate scanned image and generate violations report.
Supports the following image sources:
    {{.appName}} image validate yourrepo/yourimage:tag
    {{.appName}} image validate path/to/yourimage.tar
`, map[string]interface{}{
			"appName": internal.ApplicationName,
		}),
		Args: cobra.ExactArgs(1),
		PreRun: func(_ *cobra.Command, _ []string) {
			saasURL := config.GetConfig(config.SaasURL)
			orgKey := config.GetConfig(config.OrgKey)
			apiID := config.GetConfig(config.CBApiID)
			apiKey := config.GetConfig(config.CBApiKey)
			if buildStep == "" {
				buildStep = config.GetConfig(config.DefaultBuildStep)
			}

			validateImageHandler = validate.NewImageValidateHandler(saasURL, orgKey, apiID, apiKey, buildStep, namespace, "")

			validateScanHandler = scan.NewScanHandler(saasURL, orgKey, apiID, apiKey, nil)
			if err := validateScanHandler.HealthCheck(); err != nil {
				bus.Publish(bus.NewErrorEvent(err))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			go handleValidate(args[0])
			terminalui.NewDisplay().DisplayEvents()
		},
	}

	validateCmd.Flags().StringVarP(
		&buildStep, "build-step", "b", "", "the build step to use for validating the image")
	validateCmd.Flags().StringVarP(
		&namespace, "namespace", "n", "", "the namespace to validate the image")
	validateCmd.PersistentFlags().BoolVar(
		&opts.ForceScan, "force", false, "trigger a force scan no matter the image is scanned or not")
	validateCmd.PersistentFlags().IntVar(
		&opts.Limit, "limit", fullTable, // set to 0 will show all rows
		"number of rows to show in the report (for table format only)")

	return validateCmd
}

func handleValidate(input string) {
	err := validate.CheckValidBuildStep(buildStep)
	if err != nil {
		bus.Publish(bus.NewErrorEvent(err))
		return
	}

	scanResult, done := actualScan(input, validateScanHandler, buildStep, namespace)
	if done {
		return
	}

	validateImageHandler.AttachImageID(scanResult.ManifestDigest, scanResult.Identifier)

	result, err := validateImageHandler.Validate()
	if err != nil {
		bus.Publish(bus.NewErrorEvent(err))
		return
	}

	if len(result) == 0 {
		bus.Publish(bus.NewEvent(bus.ValidateFinishedSuccessfully,
			fmt.Sprintf("Validate results for %s finished successfully with no violations", input),
			true))

		return
	}

	bus.Publish(bus.NewEvent(
		bus.ValidateFinishedWithViolations,
		presenter.NewPresenter(
			image.NewValidatedImage(scanResult.Identifier, result, image.ValidatedImageOption{
				Option: tabletool.Option{Limit: opts.Limit},
			}),
			opts.presenterOption),
		false))

	err = cberr.NewError(cberr.PolicyViolationErr, "Validate finished with violations", nil)
	bus.Publish(bus.NewErrorEvent(err))
}
