/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bom

import (
	"github.com/anchore/syft/syft/distro"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/source"
)

// JSONDocument represents the syft cataloging findings as a JSON document.
type JSONDocument struct {
	// Artifacts is the list of packages discovered and placed into the catalog
	Artifacts             []JSONPackage      `json:"artifacts"`
	ArtifactRelationships []JSONRelationship `json:"artifactRelationships"`
	// Source represents the original object that was cataloged
	Source JSONSource `json:"source"`
	// Distro represents the Linux distribution that was detected from the source
	Distro JSONDistribution `json:"distro"`
}

// NewJSONDocument creates and populates a new JSON document struct from the given cataloging results.
func NewJSONDocument(
	catalog *pkg.Catalog, srcMetadata source.Metadata, d *distro.Distro, scope source.Scope,
) (JSONDocument, error) {
	src, err := newJSONSource(srcMetadata, scope)
	if err != nil {
		return JSONDocument{}, err
	}

	return JSONDocument{
		Artifacts:             newJSONPackages(catalog),
		ArtifactRelationships: newJSONRelationships(pkg.NewRelationships(catalog)),
		Source:                src,
		Distro:                newJSONDistribution(d),
	}, nil
}
