/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package image

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/printtool"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/image"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/presenter"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/scan"
)

var scanHandler *scan.Handler

// ScanCmd will return the image scan command.
func ScanCmd() *cobra.Command {
	scanCmd := &cobra.Command{
		Use:   "scan <source>",
		Short: "Scan an image and generate vulnerability report",
		Long: printtool.Tprintf(`Scan an image and generate vulnerability report.
Supports the following image sources:
    {{.appName}} image scan yourrepo/yourimage:tag
    {{.appName}} image scan path/to/yourimage.tar
`, map[string]interface{}{
			"appName": internal.ApplicationName,
		}),
		Args: cobra.ExactArgs(1),
		PreRun: func(_ *cobra.Command, _ []string) {
			saasURL := config.GetConfig(config.SaasURL)
			orgKey := config.GetConfig(config.OrgKey)
			apiID := config.GetConfig(config.CBApiID)
			apiKey := config.GetConfig(config.CBApiKey)

			scanHandler = scan.NewScanHandler(saasURL, orgKey, apiID, apiKey, nil)
			if err := scanHandler.HealthCheck(); err != nil {
				bus.Publish(bus.NewErrorEvent(err))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			go handleScan(args[0])
			terminalui.NewDisplay().DisplayEvents()
		},
	}

	return scanCmd
}

func handleScan(input string) {
	result, done := actualScan(input, scanHandler, "", "")
	if done {
		return
	}

	bus.Publish(bus.NewEvent(bus.ScanFinished, presenter.NewPresenter(result, opts.presenterOption), true))
}

func actualScan(input string, handler *scan.Handler, buildStep, namespace string) (*image.ScannedImage, bool) {
	var msg string

	registryHandler := scan.NewRegistryHandler()

	generatedBom, err := registryHandler.Generate(input, opts.scanOption)
	if err != nil {
		bus.Publish(bus.NewErrorEvent(err))
		return nil, true
	}

	if generatedBom == nil {
		msg = fmt.Sprintf("Generated sbom for %s is empty", input)
		e := cberr.NewError(cberr.SBOMGenerationErr, msg, err)
		bus.Publish(bus.NewErrorEvent(e))
		logrus.Errorln(e)

		return nil, true
	}

	if opts.ShouldCleanup {
		defer func() {
			// delete docker image by docker client
			if dockerClient, creationErr := client.NewClientWithOpts(); creationErr == nil {
				_, _ = dockerClient.ImageRemove(context.Background(), input, types.ImageRemoveOptions{})
			}
		}()
	}

	handler.AttachSBOMBuildStepAndNamespace(generatedBom, buildStep, namespace)

	result, err := handler.Scan(opts.scanOption)
	if err != nil {
		bus.Publish(bus.NewErrorEvent(err))
		return nil, true
	}

	return result, false
}
