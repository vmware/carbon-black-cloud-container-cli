package bom

import (
	"strings"

	"github.com/anchore/syft/syft/linux"
)

// JSONDistribution provides information about a detected Linux JSONDistribution.
type JSONDistribution struct {
	Name    string `json:"name"`    // Name of the Linux distribution
	Version string `json:"version"` // Version of the Linux distribution (major or major.minor version)
	IDLike  string `json:"idLike"`  // the ID_LIKE field found within the /etc/os-release file
}

// newJSONDistribution creates a struct with the Linux distribution to be represented in JSON.
func newJSONDistribution(d *linux.Release) JSONDistribution {
	if d == nil {
		return JSONDistribution{}
	}

	return JSONDistribution{
		Name:    d.ID,
		Version: d.VersionID,
		IDLike:  strings.Join(d.IDLike, " "),
	}
}
