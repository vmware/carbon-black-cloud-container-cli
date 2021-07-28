/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bom

import "github.com/anchore/syft/syft/distro"

// JSONDistribution provides information about a detected Linux JSONDistribution.
type JSONDistribution struct {
	Name    string `json:"name"`    // Name of the Linux distribution
	Version string `json:"version"` // Version of the Linux distribution (major or major.minor version)
	IDLike  string `json:"idLike"`  // the ID_LIKE field found within the /etc/os-release file
}

// newJSONDistribution creates a struct with the Linux distribution to be represented in JSON.
func newJSONDistribution(d *distro.Distro) JSONDistribution {
	if d == nil {
		return JSONDistribution{}
	}

	return JSONDistribution{
		Name:    d.Name(),
		Version: d.FullVersion(),
		IDLike:  d.IDLike,
	}
}
