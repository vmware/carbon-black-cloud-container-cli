package image

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.bit9.local/octarine/cbctl/internal"
	"gitlab.bit9.local/octarine/cbctl/internal/bus"
	"gitlab.bit9.local/octarine/cbctl/internal/terminalui"
	"gitlab.bit9.local/octarine/cbctl/internal/util/printtool"
	"gitlab.bit9.local/octarine/cbctl/pkg/cberr"
	"gitlab.bit9.local/octarine/cbctl/pkg/model/image"
	"gitlab.bit9.local/octarine/cbctl/pkg/presenter"
	"gitlab.bit9.local/octarine/cbctl/pkg/scan"
)

// PackagesCmd will print the image sbom.
func PackagesCmd() *cobra.Command {
	packagesCmd := &cobra.Command{
		Use:   "packages <source>",
		Short: "Print image packages",
		Long: printtool.Tprintf(`Download an image and print the image packages:
    {{.appName}} image packages yourrepo/yourimage:tag
    {{.appName}} image packages path/to/yourimage.tar
`, map[string]interface{}{
			"appName": internal.ApplicationName,
		}),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			go PrintSBOM(args[0])
			terminalui.NewDisplay().DisplayEvents()
		},
	}

	return packagesCmd
}

// PrintSBOM will print the image SBOM.
func PrintSBOM(input string) {
	var msg string

	registryHandler := scan.NewRegistryHandler()

	generatedBom, err := registryHandler.Generate(input, opts.scanOption)
	if err != nil {
		bus.Publish(bus.NewErrorEvent(err))
	}

	if generatedBom == nil {
		msg = fmt.Sprintf("Generated packages for %s is empty", input)
		e := cberr.NewError(cberr.SBOMGenerationErr, msg, err)
		bus.Publish(bus.NewErrorEvent(e))
		logrus.Errorln(e)

		return
	}

	sbomImage := image.SBOM{
		FullTag:        generatedBom.FullTag,
		ManifestDigest: generatedBom.ManifestDigest,
		Packages:       generatedBom.Packages,
	}

	opts.presenterOption.Limit = len(generatedBom.Packages.Artifacts)
	bus.Publish(bus.NewEvent(bus.PrintSBOM, presenter.NewPresenter(&sbomImage, opts.presenterOption), true))
}
