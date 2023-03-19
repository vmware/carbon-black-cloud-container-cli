package image

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/version"
	"strings"

	progress "github.com/wagoodman/go-progress"

	containersimage "github.com/containers/image/v5/image"
	"github.com/containers/image/v5/transports/alltransports"
	imagetype "github.com/containers/image/v5/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/internal"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/terminalui"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/printtool"
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

			scanHandler = scan.NewScanHandler(saasURL, orgKey, apiID, apiKey, nil, nil)
			if err := scanHandler.HealthCheck(); err != nil {
				bus.Publish(bus.NewErrorEvent(err))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			go handleScan(args[0])
			terminalui.NewDisplay().DisplayEvents()
		},
	}

	scanCmd.PersistentFlags().BoolVar(
		&opts.ForceScan, "force", false, "trigger a force scan no matter the image is scanned or not")
	scanCmd.PersistentFlags().IntVar(
		&opts.Limit, "limit", fullTable, // set to 0 will show all rows
		"number of rows to show in the report (for table format only)")

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
	stage := &progress.Stage{Current: "Fetch image id"}
	prog := &progress.Manual{}
	prog.SetTotal(1)
	value := progress.StagedProgressable(&struct {
		progress.Stager
		progress.Progressable
	}{
		Stager:       stage,
		Progressable: prog,
	})
	bus.Publish(bus.NewEvent(bus.StartScanTryFetchImageID, value, false))
	defer prog.SetCompleted()

	operationID := uuid.New().String()
	logrus.WithField("operation_id", operationID).Info("Starting an operation")

	imageID, err := getImageID(input)
	if imageID != "" && !opts.ForceScan && opts.presenterOption.OutputFormat != "cyclondx" {
		if err == nil {
			versionInfo := version.GetCurrentVersion()
			results, err := handler.GetImagesScanResultsFromBackendByImageID(imageID, versionInfo.Version)
			if err == nil {
				return results, false
			}
		}
	}

	stage.Current = "Done fetching image id"

	scanner := scan.NewScanner()
	generatedBom, imgLayers, hasErr := scanner.ExtractDataFromImage(input, opts.scanOption)
	if hasErr {
		return nil, true
	}

	handler.AttachData(generatedBom, imgLayers, buildStep, namespace, imageID)

	result, err := handler.Scan(operationID, opts.scanOption)
	if err != nil {
		bus.Publish(bus.NewErrorEvent(err))
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

	return result, false
}

func getImageID(input string) (string, error) {
	ctx := context.Background()
	srcCtx := &imagetype.SystemContext{
		// if a multi-arch image detected, pull the linux image by default
		ArchitectureChoice:          "amd64",
		OSChoice:                    "linux",
		DockerInsecureSkipTLSVerify: imagetype.OptionalBoolTrue,
	}

	src, err := parseImageSource(ctx, srcCtx, input)
	if err != nil {
		return "", err
	}

	defer func(src imagetype.ImageSource) {
		_ = src.Close()
	}(src)

	img, err := containersimage.FromUnparsedImage(ctx, srcCtx, containersimage.UnparsedInstance(src, nil))
	if err != nil {
		return "", err
	}

	configBlob, err := img.ConfigBlob(ctx)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(configBlob)

	configDigest := hex.EncodeToString(hash[:])
	if configDigest == "" {
		return "", fmt.Errorf("empty image id")
	}

	configDigest = "sha256:" + configDigest

	return configDigest, nil
}

// parseImageSource converts image URL-like string to an ImageSource.
// The caller must call .Close() on the returned ImageSource.
func parseImageSource(ctx context.Context, srcCtx *imagetype.SystemContext, name string) (imagetype.ImageSource, error) {
	transport := alltransports.TransportFromImageName(name)
	if transport == nil && !strings.Contains(name, ".tar") {
		name = "docker://" + name
	}

	ref, err := alltransports.ParseImageName(name)
	if err != nil {
		return nil, err
	}

	return ref.NewImageSource(ctx, srcCtx)
}
