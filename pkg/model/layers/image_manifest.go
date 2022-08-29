package layers

import (
	"encoding/json"
)

type Manifest struct {
	ConfigPath    string   `json:"Config"`
	RepoTags      []string `json:"RepoTags"`
	LayerTarPaths []string `json:"Layers"`
}

func NewManifest(manifestBytes []byte) (*Manifest, error) {
	var manifest []Manifest
	err := json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		return nil, err
	}
	return &manifest[0], nil
}
