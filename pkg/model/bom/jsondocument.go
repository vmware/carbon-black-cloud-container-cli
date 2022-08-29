package bom

import (
	"github.com/anchore/syft/syft/linux"
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
func NewJSONDocument(catalog *pkg.Catalog, srcMetadata source.Metadata, d *linux.Release, scope source.Scope,
) (JSONDocument, error) {
	src, err := newJSONSource(srcMetadata, scope)
	if err != nil {
		return JSONDocument{}, err
	}

	// we put an empty catalog into the relationships filed because it is providing the best scan results.
	// we need to check this again when we upgrade Ancore next time.
	return JSONDocument{
		Artifacts:             newJSONPackages(catalog),
		ArtifactRelationships: newJSONRelationships(pkg.NewRelationships(pkg.NewCatalog())),
		Source:                src,
		Distro:                newJSONDistribution(d),
	}, nil
}
