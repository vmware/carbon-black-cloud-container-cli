package image

import "encoding/xml"

// source: https://github.com/CycloneDX/specification

// Document represents a CycloneDX VulnerabilityCyclon Document.
type Document struct {
	XMLName       xml.Name       `xml:"bom"`
	XMLNs         string         `xml:"xmlns,attr"`
	XMLNsV        string         `xml:"xmlns:v,attr"`
	Version       int            `xml:"version,attr"`
	SerialNumber  string         `xml:"serialNumber,attr"`
	BomDescriptor *BomDescriptor `xml:"metadata"`
	Components    []Component    `xml:"components>component"`
}
