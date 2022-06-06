package image

import (
	"encoding/xml"
	"time"
)

// Source: https://cyclonedx.org/ext/bom-descriptor/

// BomDescriptor represents all metadata surrounding the BOM report (such as when the BOM was made, with which tool,
// and the item being cataloged).
type BomDescriptor struct {
	XMLName   xml.Name     `xml:"metadata"`
	Timestamp string       `xml:"timestamp,omitempty"` // The date and time (timestamp) when the document was created
	Tools     []BdTool     `xml:"tools>tool"`          // The tool used to create the BOM.
	Component *BdComponent `xml:"component"`           // The Component that the BOM describes.
}

// BdTool represents the tool that created the BOM report.
type BdTool struct {
	XMLName xml.Name `xml:"tool"`
	Vendor  string   `xml:"vendor,omitempty"` // The vendor of the tool used to create the BOM.
	Name    string   `xml:"name,omitempty"`   // The name of the tool used to create the BOM.
	Version string   `xml:"version"`          // The version of the tool used to create the BOM.
}

// BdComponent represents the software/package being cataloged.
type BdComponent struct {
	XMLName xml.Name `xml:"component"`
	Component
}

// NewBomDescriptor returns a new BomDescriptor tailored for the current time and "syft" tool details.
func NewBomDescriptor(name, version string, userInput string, digest string) *BomDescriptor {
	descriptor := BomDescriptor{
		XMLName:   xml.Name{},
		Timestamp: time.Now().Format(time.RFC3339),
		Tools: []BdTool{
			{
				Vendor:  "VMware",
				Name:    name,
				Version: version,
			},
		},
	}

	descriptor.Component = &BdComponent{
		Component: Component{
			Type:    "container",
			Name:    userInput,
			Version: digest,
		},
	}

	return &descriptor
}
