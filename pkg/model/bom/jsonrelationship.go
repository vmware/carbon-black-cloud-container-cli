package bom

import (
	"github.com/anchore/syft/syft/artifact"
)

// JSONRelationship denotes the relationship of artifacts.
type JSONRelationship struct {
	Parent   string      `json:"parent"`
	Child    string      `json:"child"`
	Type     string      `json:"type"`
	Metadata interface{} `json:"metadata"`
}

func newJSONRelationships(relationships []artifact.Relationship) []JSONRelationship {
	result := make([]JSONRelationship, len(relationships))
	for i, r := range relationships {
		result[i] = JSONRelationship{
			Parent:   string(r.From.ID()),
			Child:    string(r.To.ID()),
			Type:     string(r.Type),
			Metadata: r.Data,
		}
	}

	return result
}
