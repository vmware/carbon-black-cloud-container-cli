package layers

import (
	"encoding/json"
)

type ImageConfig struct {
	History []historyEntry `json:"history"`
	RootFs  rootFs         `json:"rootfs"`
}

type rootFs struct {
	Type    string   `json:"type"`
	DiffIds []string `json:"diff_ids"`
}

type historyEntry struct {
	ID         string
	Size       uint64
	Created    string `json:"created"`
	Author     string `json:"author"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer"`
}

func NewImageConfig(configBytes []byte) (*ImageConfig, error) {
	var imageConfig ImageConfig
	err := json.Unmarshal(configBytes, &imageConfig)
	if err != nil {
		return nil, err
	}

	layerIdx := 0
	for idx := range imageConfig.History {
		currentLayer := &imageConfig.History[idx]
		if currentLayer.EmptyLayer {
			currentLayer.ID = ""
		} else {
			currentLayer.ID = imageConfig.RootFs.DiffIds[layerIdx]
			layerIdx++
		}
	}

	return &imageConfig, err
}
