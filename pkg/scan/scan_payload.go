package scan

import (
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/bom"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/layers"
	"strconv"
)

const (
	layerDigest     = "LAYER_DIGEST"
	layerCommand    = "COMMAND"
	layerSize       = "SIZE"
	layerFilesCount = "FILES_COUNT"
)

// AnalysisPayload is the payload used for uploading sbom to image scanning service.
type AnalysisPayload struct {
	SBOM      *bom.JSONDocument `json:"sbom"`
	Layers    []layers.Layer    `json:"layers"`
	BuildStep string            `json:"build_step"`
	Namespace string            `json:"namespace"`
	ForceScan bool              `json:"force_scan"`
	ImageID   string            `json:"image_id"`
	Meta      struct {
		SyftVersion string `json:"syft_version"`
		CliVersion  string `json:"cli_version"`
	} `json:"metadata"`
}

func NewAnalysisPayload(sbom *bom.JSONDocument, layers []layers.Layer, buildStep, namespace string, forceScan bool, syftVersion, cliVersion string) AnalysisPayload {
	return AnalysisPayload{
		SBOM:      sbom,
		Layers:    layers,
		BuildStep: buildStep,
		Namespace: namespace,
		ForceScan: forceScan,
		//ImageID:   imageID,
		Meta: struct {
			SyftVersion string `json:"syft_version"`
			CliVersion  string `json:"cli_version"`
		}{
			SyftVersion: syftVersion,
			CliVersion:  cliVersion,
		},
	}
}

func (payload AnalysisPayload) Title() string {
	return ""
}

func (payload AnalysisPayload) Footer() string {
	return ""
}

func (payload AnalysisPayload) Header() []string {
	return []string{
		layerDigest,
		layerCommand,
		layerSize,
		layerFilesCount,
	}
}

func (payload AnalysisPayload) Rows() [][]string {
	result := make([][]string, 0)

	for _, layer := range payload.Layers {
		result = append(result, []string{
			layer.Digest,
			layer.Command,
			strconv.FormatUint(layer.Size, 10),
			strconv.Itoa(len(layer.Files)),
		})
	}

	return result
}
