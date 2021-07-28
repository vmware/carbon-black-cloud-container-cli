/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bom

import (
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/source"
)

// JSONPackage represents a pkg.Package object specialized for JSON marshaling and unmarshaling.
type JSONPackage struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Type         string            `json:"type"`
	FoundBy      string            `json:"foundBy"`
	Locations    []source.Location `json:"locations"`
	Licenses     []string          `json:"licenses"`
	Language     string            `json:"language"`
	CPEs         []string          `json:"cpes"`
	PURL         string            `json:"purl"`
	MetadataType string            `json:"metadataType"`
	Metadata     interface{}       `json:"metadata"`
}

func newJSONPackages(catalog *pkg.Catalog) []JSONPackage {
	artifacts := make([]JSONPackage, 0)

	if catalog == nil {
		return artifacts
	}

	for _, p := range catalog.Sorted() {
		art := newJSONPackage(p)
		artifacts = append(artifacts, art)
	}

	return artifacts
}

// newJSONPackage crates a new JSONPackage from the given pkg.Package.
func newJSONPackage(p *pkg.Package) JSONPackage {
	cpes := make([]string, len(p.CPEs))
	for i, c := range p.CPEs {
		cpes[i] = c.BindToFmtString()
	}

	// ensure collections are never nil for presentation reasons
	locations := make([]source.Location, 0)
	if p.Locations != nil {
		locations = p.Locations
	}

	licenses := make([]string, 0)
	if p.Licenses != nil {
		licenses = p.Licenses
	}

	return JSONPackage{
		ID:           string(p.ID),
		Name:         p.Name,
		Version:      p.Version,
		Type:         string(p.Type),
		FoundBy:      p.FoundBy,
		Locations:    locations,
		Licenses:     licenses,
		Language:     string(p.Language),
		CPEs:         cpes,
		PURL:         p.PURL,
		MetadataType: string(p.MetadataType),
		Metadata:     p.Metadata,
	}
}
