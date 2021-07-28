/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bom

import "github.com/anchore/syft/syft/pkg"

// JSONRelationship denotes the relationship of artifacts.
type JSONRelationship struct {
	Parent   string      `json:"parent"`
	Child    string      `json:"child"`
	Type     string      `json:"type"`
	Metadata interface{} `json:"metadata"`
}

func newJSONRelationships(relationships []pkg.Relationship) []JSONRelationship {
	result := make([]JSONRelationship, len(relationships))
	for i, r := range relationships {
		result[i] = JSONRelationship{
			Parent:   string(r.Parent),
			Child:    string(r.Child),
			Type:     string(r.Type),
			Metadata: r.Metadata,
		}
	}

	return result
}
