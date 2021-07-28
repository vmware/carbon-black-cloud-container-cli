/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package image

import (
	"fmt"
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
