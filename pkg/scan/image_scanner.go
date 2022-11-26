package scan

import (
	"fmt"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/bus"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/cberr"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/layers"
	progress "github.com/wagoodman/go-progress"
	"sync"
)

type Scanner struct{}

// NewScanner creates a new Scanner that captures all supported scan operations under one interface
func NewScanner() *Scanner {
	return &Scanner{}
}

// GenerateSBOM is a wrapper around scan.GenerateSBOMFromImage
func (s *Scanner) GenerateSBOM(img *image.Image, userInput string, opts Option) (*Bom, error) {
	// Note: progress and events are handled by syft internally so we don't raise any events here
	return GenerateSBOMFromImage(img, userInput, opts.FullTag)
}

// GenerateLayersAndFiles is a wrapper around scan.GenerateLayersAndFileData
func (s *Scanner) GenerateLayersAndFiles(img *image.Image, _ string, _ Option) ([]layers.Layer, error) {
	stage := &progress.Stage{Current: "Reading layers from image"}
	prog := &progress.Manual{Total: 1}
	value := progress.StagedProgressable(&struct {
		progress.Stager
		progress.Progressable
	}{
		Stager:       stage,
		Progressable: prog,
	})
	bus.Publish(bus.NewEvent(bus.NewCollectLayers, value, false))
	defer prog.SetCompleted()

	foundLayers, err := GenerateLayersAndFileData(img)
	if err != nil {
		stage.Current = "failed"
		return nil, err
	}

	stage.Current = fmt.Sprintf("%d layers", len(foundLayers))
	return foundLayers, nil
}

func (s *Scanner) ExtractDataFromImage(input string, opts Option) (*Bom, []layers.Layer, bool) {
	var msg string
	registryHandler := NewRegistryHandler()

	img, err := registryHandler.LoadImage(input, opts)
	if err != nil {
		msg := fmt.Sprintf("Failed to pull image for input %s", input)
		e := cberr.NewError(cberr.ImageLoadErr, msg, err)
		bus.Publish(bus.NewErrorEvent(e))
		logrus.Errorln(e)
		return nil, nil, true
	}
	defer func() {
		if err := img.Cleanup(); err != nil {
			logrus.WithError(err).Errorf("failed to clean up files for image [%s]", input)
		}
		Cleanup()
	}()

	var generatedBom *Bom
	var imgLayers []layers.Layer
	var errBom, errLayers error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		imgLayers, errLayers = s.GenerateLayersAndFiles(img, input, opts)
		wg.Done()
	}()

	go func() {
		generatedBom, errBom = s.GenerateSBOM(img, input, opts)
		wg.Done()
	}()

	wg.Wait()

	if errBom != nil {
		bus.Publish(bus.NewErrorEvent(errBom))
		return nil, nil, true
	}

	if generatedBom == nil {
		msg = fmt.Sprintf("Generated sbom for %s is empty", input)
		e := cberr.NewError(cberr.SBOMGenerationErr, msg, errBom)
		bus.Publish(bus.NewErrorEvent(e))
		logrus.Errorln(e)
		return nil, nil, true
	}

	if errLayers != nil {
		bus.Publish(bus.NewErrorEvent(errLayers))
		return nil, nil, true
	}

	if imgLayers == nil {
		msg := fmt.Sprintf("No layers were found for input %s", input)
		e := cberr.NewError(cberr.LayersGenerationErr, msg, nil)
		bus.Publish(bus.NewErrorEvent(e))
		logrus.Errorln(e)
		return nil, nil, true
	}

	return generatedBom, imgLayers, false
}
