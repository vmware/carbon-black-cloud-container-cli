package k8sobject

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/config"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
	"gitlab.bit9.local/octarine/cbctl/pkg/presenter"
	"gitlab.bit9.local/octarine/cbctl/pkg/validate"
)

var (
	// Flag options for validation.
	buildStep string
	namespace string
	path      string

	validateResourceHandler *validate.K8SObjectHandler
)

// ValidateCmd will return the image validate command.
func ValidateCmd() *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate -f path",
		Short: "Validate k8s resource/s and generate violations report",
		Long:  "Validate k8s resource/s and generate violations report",
		Args:  cobra.NoArgs,
		PreRun: func(_ *cobra.Command, _ []string) {
			if path == "" {
				e := cberr.NewError(cberr.ConfigErr, "Must specify the resource argument -f", nil)
				bus.Publish(bus.NewErrorEvent(e))
			}

			saasURL := config.GetConfig(config.SaasURL)
			orgKey := config.GetConfig(config.OrgKey)
			apiID := config.GetConfig(config.CBApiID)
			apiKey := config.GetConfig(config.CBApiKey)
			if buildStep == "" {
				buildStep = config.GetConfig(config.DefaultBuildStep)
			}

			validateResourceHandler = validate.NewK8SObjectValidateHandler(
				saasURL, orgKey, apiID, apiKey, buildStep, namespace, path)
		},
		Run: func(cmd *cobra.Command, args []string) {
			go handleValidate()
			terminalui.NewDisplay().DisplayEvents()
		},
	}

	validateCmd.Flags().StringVarP(
		&buildStep, "build-step", "b", "", "the build step to use for validating the image")
	validateCmd.Flags().StringVarP(
		&namespace, "namespace", "n", "", "the namespace to validate the image")
	validateCmd.Flags().StringVarP(
		&path, "file", "f", "", "the value for resources' path")

	return validateCmd
}

func handleValidate() {
	err := validate.CheckValidBuildStep(buildStep)
	if err != nil {
		bus.Publish(bus.NewErrorEvent(err))
		return
	}

	result, err := validateResourceHandler.Validate()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to validate k8s-resource: %s", err.Error())
		bus.Publish(bus.NewErrorEvent(cberr.NewError(cberr.ValidateFailedErr, errMsg, err)))

		return
	}

	if err = result.GetErrors(); err != nil {
		errMsg := fmt.Sprintf("Validate k8s-object finished. %s", err.Error())
		bus.Publish(bus.NewErrorEvent(cberr.NewError(cberr.ValidateFailedErr, errMsg, err)))

		return
	}

	resultByPolicy := result.ToValidatedResourcesByPolicy()
	if resultByPolicy.PolicyViolationsCount() == 0 {
		bus.Publish(bus.NewEvent(bus.ValidateFinishedSuccessfully,
			fmt.Sprintf("Validate results for %s finished successfully with no violations", path),
			true))

		return
	}

	for _, policyViolatingResources := range resultByPolicy {
		bus.Publish(bus.NewEvent(
			bus.ValidateFinishedWithViolations,
			presenter.NewPresenter(policyViolatingResources, opts.presenterOption),
			false))
	}

	err = cberr.NewError(cberr.PolicyViolationErr, "Validate finished with violations", nil)
	bus.Publish(bus.NewErrorEvent(err))
}
