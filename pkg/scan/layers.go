package scan

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/layers"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const manifestFileName = "manifest.json"

func GenerateLayers(manifestDigest string, tarPath string, configBytes []byte) ([]layers.Layer, error) {
	imageConfig, err := layers.NewImageConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing image config: %w", err)
	}

	tarFile, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer tarFile.Close()

	layersMap := make(map[string]layers.Layer)
	tarReader := tar.NewReader(tarFile)

	var imageManifest *layers.Manifest
	var currentLayer uint
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := header.Name

		// some layer tars can be relative layer symlinks to other layer tars
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeReg {

			if strings.HasSuffix(name, ".tar") {
				currentLayer++
				// add tar reader
				layerReader := tar.NewReader(tarReader)
				layer, err := processLayerTar(layerReader) // if we won't find digest in the history, we will use the tar name
				if err != nil {
					return nil, err
				}
				layersMap[name] = layer

			} else if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, "tgz") {
				currentLayer++
				// add gzip reader
				gz, err := gzip.NewReader(tarReader)
				if err != nil {
					return nil, err
				}
				// add tar reader
				layerReader := tar.NewReader(gz)
				layer, err := processLayerTar(layerReader)
				if err != nil {
					return nil, err
				}
				layersMap[name] = layer

			} else if name == manifestFileName {
				manifestBytes, err := ioutil.ReadAll(tarReader)
				if err != nil {
					return nil, fmt.Errorf("error parsing image manifest: %v", err)
				}
				manifest, err := layers.NewManifest(manifestBytes)
				imageManifest = manifest
			}
		}
	}

	if imageManifest == nil {
		return nil, fmt.Errorf("could not find image manifest")
	}

	return parseLayers(manifestDigest, layersMap, *imageManifest, *imageConfig)
}

func processLayerTar(tarReader *tar.Reader) (layers.Layer, error) {
	var files []layers.ExecutableFile
	layer := layers.Layer{
		Size: 0,
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return layer, err
		}

		// always ensure relative path notations are not parsed as part of the filename
		name := path.Clean(header.Name)
		if name == "." {
			continue
		}

		switch header.Typeflag {
		case tar.TypeXGlobalHeader:
			return layer, fmt.Errorf("unexptected tar file: (XGlobalHeader): type=%v name=%s", header.Typeflag, name)
		case tar.TypeXHeader:
			return layer, fmt.Errorf("unexptected tar file (XHeader): type=%v name=%s", header.Typeflag, name)
		default:
			// sum the size of the files
			layer.Size += uint64(header.Size)
			file, err := layers.ReadExecutableFromTar(tarReader, header, name)
			if err != nil {
				return layer, err
			}
			if file != nil {
				files = append(files, *file)
			}
		}
	}
	layer.Files = files
	return layer, nil
}

func parseLayers(manifestDigest string, layersMap map[string]layers.Layer, manifest layers.Manifest, config layers.ImageConfig) ([]layers.Layer, error) {
	nonEmptyLayers := make([]layers.Layer, 0)
	// get the relevant imageLayers and sort them by the manifest
	for _, layerName := range manifest.LayerTarPaths {
		layer, exists := layersMap[layerName]
		if exists {
			nonEmptyLayers = append(nonEmptyLayers, layer)
			continue
		}
		return nil, fmt.Errorf("could not find '%s' in parsed imageLayers", layerName)
	}

	// build the layers list by the config
	imageLayers := make([]layers.Layer, 0)
	layerIdx := 0
	nonEmptyLayersIdx := 0
	for _, history := range config.History {
		if !history.EmptyLayer {
			layer := nonEmptyLayers[nonEmptyLayersIdx]
			layer.Digest = history.ID
			layer.Command = strings.TrimPrefix(history.CreatedBy, "/bin/sh -c ")
			layer.Index = layerIdx

			imageLayers = append(imageLayers, layer)
			nonEmptyLayersIdx++
		} else {
			layer := layers.Layer{
				// generating an empty layer digest because it doesn't have a digest
				Digest:  fmt.Sprintf("%v_%v", manifestDigest, layerIdx),
				Command: strings.TrimPrefix(history.CreatedBy, "/bin/sh -c "),
				Index:   layerIdx,
			}
			imageLayers = append(imageLayers, layer)
		}
		layerIdx++
	}

	return imageLayers, nil
}
