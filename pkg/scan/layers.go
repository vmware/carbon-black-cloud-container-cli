package scan

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/layers"
)

// GenerateLayersAndFileData reads the input image and calculates:
// - metadata for all layers (including empty ones) in the image
// - interesting files per layer with an flag whether they are in the squashed image or not
func GenerateLayersAndFileData(image *image.Image) ([]layers.Layer, error) {
	if image == nil {
		return nil, errors.New("image for layers extraction can't be nil")
	}

	logrus.Info("Starting to read layers and files from the image")
	convertedLayers := make([]layers.Layer, 0, len(image.Metadata.Config.History))
	manifestDigest := fmt.Sprintf("%x", sha256.Sum256(image.Metadata.RawManifest))

	indexlayers := 0
	for ixHistory, historyEntry := range image.Metadata.Config.History {
		logrus.Debugf("Reading layer num %d", ixHistory)
		convertedLayer := layers.Layer{
			Command: historyEntry.CreatedBy,
			Size:    0,
			Index:   ixHistory,
			IsEmpty: historyEntry.EmptyLayer,
		}

		if historyEntry.EmptyLayer {
			// Layer that does not change the filesystem; we use a custom digest for unique purposes
			convertedLayer.Digest = fmt.Sprintf("sha256:%s_%d", manifestDigest, ixHistory)
		} else {
			// Find the matching "real" layer with changes to the filesystem
			layer := image.Layers[indexlayers]
			indexlayers++

			convertedLayer.Digest = layer.Metadata.Digest
			convertedLayer.Size = uint64(layer.Metadata.Size)

			convertedFilesForLayer := make([]layers.ExecutableFile, 0)
			for _, fileRef := range layer.Tree.AllFiles() {
				convertedFile, err := readFileForLayer(fileRef, image, layer)
				if err != nil {
					// TODO: Stop skipping this error once file analysis is in GA?
					logrus.WithError(err).Errorf("reading file (%v) in layer (%v) failed and file could not be analyzed", fileRef.RealPath, layer.Metadata.Digest)
					continue
				}
				if convertedFile != nil {
					convertedFilesForLayer = append(convertedFilesForLayer, *convertedFile)
				}
			}
			convertedLayer.Files = convertedFilesForLayer
		}

		logrus.Debugf("Read layer num %d and digest %v. Found %d files", ixHistory, convertedLayer.Digest, len(convertedLayer.Files))
		convertedLayers = append(convertedLayers, convertedLayer)
	}

	logrus.Info("Finished reading layers and files from the image")
	return convertedLayers, nil
}

// readFileForLayer reads a specific file ref from the image's layer and decides if it should be processed or not
// the first return value is the file meta, if the file is interesting and should be processed, or nil otherwise
func readFileForLayer(fileRef file.Reference, img *image.Image, layer *image.Layer) (*layers.ExecutableFile, error) {
	fRead, err := layer.FileContents(fileRef.RealPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := fRead.Close(); err != nil {
			logrus.WithError(err).Errorf("failed to close stream reader for file (%v) from image archive in layer (%v)", fileRef.RealPath, layer.Metadata.Digest)
		}
	}()

	isELF, hash, err := layers.CalculateELFMetadata(fRead)
	if isELF {
		fileMeta, err := img.FileCatalog.Get(fileRef)
		if err != nil {
			return nil, err
		}

		convertedFile := layers.ExecutableFile{
			Digest:   hash,
			Path:     fileMeta.Metadata.Path,
			Size:     uint64(fileMeta.Metadata.Size),
			Category: layers.CategoryElf,
		}

		// Find if the file is in the final squashed tree and append in that list if needed
		inSquashedStereo, squashedRef, err := img.SquashedTree().File(fileRef.RealPath)
		if err != nil {
			return nil, err
		}

		// Validates that a file exists at this path in the final image (inSquashedStereo)
		// AND
		// the file came from the current layer
		// If inSquashedStereo=true but the ref's ID is different, then the file was modified in a later layer
		// And the current layer's version is NOT the one in the final image
		// This is because stereoscope ensures each new iteration of a file has a new ID() across all layers
		// squashedRef should never be nil if inSquashedStereo=true but better safe than sorry
		convertedFile.InSquashedImage = inSquashedStereo && squashedRef != nil && squashedRef.ID() == fileRef.ID()

		return &convertedFile, nil
	}

	return nil, nil
}
