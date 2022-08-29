package image

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.bit9.local/octarine/cbctl/internal/version"
	"gitlab.bit9.local/octarine/cbctl/pkg/model/bom"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	vulnerabilityHeader = "Vuln ID"
	packageHeader       = "Package"
	typeHeader          = "Type"
	severityHeader      = "Severity"
	fixAvailableHeader  = "Fix Available"
	cvssV2Header        = "CVSS V2"
	cvssV3Header        = "CVSS V3"
)

// ScannedImage response model from image scanning service.
type ScannedImage struct {
	Identifier       `json:",inline"`
	ImageMetadata    Metadata          `json:"image_metadata"`
	Account          string            `json:"account"`
	ScanStatus       string            `json:"scan_status"`
	Vulnerabilities  []Vulnerability   `json:"vulnerabilities"`
	PolicyViolations []PolicyViolation `json:"policy_violations,omitempty"`
	Packages         bom.JSONDocument  `json:"packages"`
}

// Title is the title of the ScannedImage result.
func (s *ScannedImage) Title() string {
	return fmt.Sprintf("Scan result for %s (%s):", s.FullTag, s.ManifestDigest)
}

// Header is the header columns of the ScannedImage result.
func (s *ScannedImage) Header() []string {
	return []string{
		vulnerabilityHeader,
		packageHeader,
		typeHeader,
		severityHeader,
		fixAvailableHeader,
		cvssV2Header,
		cvssV3Header,
	}
}

// Rows returns all the vulnerabilities of the ScannedImage result as list of rows.
func (s *ScannedImage) Rows() [][]string {
	result := make([][]string, 0)

	sortVulnerabilitiesBySeverities(s.Vulnerabilities)

	for _, vul := range s.Vulnerabilities {
		result = append(result, []string{
			vul.GetID(),
			vul.GetPackage(),
			vul.GetType(),
			vul.GetSeverity(),
			vul.GetFixAvailable(),
			vul.GetCvssV2(),
			vul.GetCvssV3(),
		})
	}

	return result
}

// CycloneDXDoc returns all the vulnerabilities of the ScannedImage result as list of rows.
func (s *ScannedImage) CycloneDXDoc() ([]byte, error) {
	doc := Document{
		XMLNs:        "http://cyclonedx.org/schema/bom/1.2",
		XMLNsV:       "http://cyclonedx.org/schema/ext/vulnerability/1.0",
		Version:      1,
		SerialNumber: uuid.New().URN(),
	}

	cbctlVersion := version.GetCurrentVersion()
	doc.BomDescriptor = NewBomDescriptor("Carbon Black", cbctlVersion.Version, s.FullTag, s.ManifestDigest)

	for _, artifact := range s.Packages.Artifacts {
		// make a new Component (by value)
		component := Component{
			Type:    "library",
			Name:    artifact.Name,
			Version: artifact.Version,
		}

		var licenses []License
		for _, licenseName := range artifact.Licenses {
			licenses = append(licenses, License{
				Name: licenseName,
			})
		}

		if len(licenses) > 0 {
			// adding licenses to the Component
			component.Licenses = &licenses
		}

		var vulnerabilities []VulnerabilityCyclon

		for _, vul := range s.Vulnerabilities {
			if artifact.Name == vul.Name {
				var vulnerability VulnerabilityCyclon
				vulnerability.ID = vul.ID
				vulnerability.Description = vul.Description
				vulnerability.Source.Name = vul.Type
				vulnerability.Source.URL = MakeVulnerabilityURL(vul.ID)

				ratings := make([]Rating, 1)
				severity := cases.Title(language.AmericanEnglish).String(strings.ToLower(vul.Severity))
				ratings[0].Severity = severity
				vulnerability.Ratings = ratings

				fixAvailable := make([]string, 1)
				fixAvailable[0] = vul.FixAvailable

				myAdvisory := new(Advisories)
				myAdvisory.Advisory = fixAvailable

				vulnerability.Advisories = myAdvisory

				vulnerabilities = append(vulnerabilities, vulnerability)
			}
		}

		if len(vulnerabilities) > 0 {
			component.Vulnerabilities = &vulnerabilities
		}

		doc.Components = append(doc.Components, component)
	}

	return xml.MarshalIndent(doc, "", " ")
}
