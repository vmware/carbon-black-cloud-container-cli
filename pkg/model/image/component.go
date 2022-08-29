package image

import (
	"encoding/xml"
)

// Component represents a single element in the CycloneDX BOM.
type Component struct {
	XMLName xml.Name `xml:"component"`
	// Required; Describes if the Component is a library, framework, application, container, operating system, firmware,
	// hardware device, or file
	Type string `xml:"type,attr"`
	// The organization that supplied the Component.
	// The supplier may often be the manufacture, but may also be a distributor or repackager.
	Supplier string `xml:"supplier,omitempty"`
	// The person(s) or organization(s) that authored the Component
	Author string `xml:"author,omitempty"`
	// The person(s) or organization(s) that published the Component
	Publisher string `xml:"publisher,omitempty"`
	// The high-level classification that a project self-describes as. This will often be a shortened,
	// single name of the company or project that produced the Component, or the source package or domain name.
	Group string `xml:"group,omitempty"`
	// Required; The name of the Component as defined by the project
	Name string `xml:"name"`
	// Required; The version of the Component as defined by the project
	Version string `xml:"version"`
	// A description of the Component
	Description string `xml:"description,omitempty"`
	// A node describing zero or more License names, SPDX License IDs or expressions
	Licenses *[]License `xml:"licenses>License"`
	// Specifies the package-url (PackageURL). The purl,
	// if specified, must be valid and conform to the specification defined at: https://github.com/package-url/purl-spec
	PackageURL      string                 `xml:"purl,omitempty"`
	Vulnerabilities *[]VulnerabilityCyclon `xml:"v:vulnerabilities>v:vulnerability,omitempty"`
}

// License represents a single software License for a Component.
type License struct {
	XMLName xml.Name `xml:"license"`
	// A valid SPDX License ID
	ID string `xml:"id,omitempty"`
	// If SPDX does not define the License used, this field may be used to provide the License name.
	Name string `xml:"name,omitempty"`
}
