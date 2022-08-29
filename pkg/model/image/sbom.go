package image

import (
	"fmt"

	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/bom"
)

const (
	nameHeader    = "NAME"
	versionHeader = "VERSION"
	bomTypeHeader = "TYPE"
)

// SBOM response model from image scanning service.
type SBOM struct {
	// FullTag is the full tag of the bom
	FullTag string
	// ManifestDigest is the sha256 of this image manifest json
	ManifestDigest string
	// Packages enumerates the packages in the bill of materials
	Packages bom.JSONDocument
}

// Title is the title of the SBOM result.
func (s *SBOM) Title() string {
	return fmt.Sprintf("Packages found for %s (%s):", s.FullTag, s.ManifestDigest)
}

// Header is the header columns of the SBOM result.
func (s *SBOM) Header() []string {
	return []string{
		nameHeader,
		versionHeader,
		bomTypeHeader,
	}
}

// Footer for adding notes in output footer.
func (s *SBOM) Footer() string {
	return ""
}

// Rows returns all the SBOM names versions and types as list of rows.
func (s *SBOM) Rows() [][]string {
	result := make([][]string, 0)

	for _, vul := range s.Packages.Artifacts {
		result = append(result, []string{
			vul.Name,
			vul.Version,
			vul.Type,
		})
	}

	return result
}
