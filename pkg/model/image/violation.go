/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package image

import (
	"bytes"
	"strings"

	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/tabletool"
)

// PolicyViolation represent a violation of a policy.
type PolicyViolation struct {
	Policy    string     `json:"policy"`
	Rule      string     `json:"rule"`
	Risk      string     `json:"risk"`
	Violation Violations `json:"violation"`
}

// GetPolicyName implements the GetPolicyName method needed for presenting policy table by the presenter.
func (p PolicyViolation) GetPolicyName() string {
	return p.Policy
}

// GetRuleName implements the GetRuleName method needed for presenting policy table by the presenter.
func (p PolicyViolation) GetRuleName() string {
	return p.Rule
}

// GetRisk implements the GetRisk method needed for presenting policy table by the presenter.
func (p PolicyViolation) GetRisk() string {
	return p.Risk
}

// GetViolation implements the GetViolation method needed for presenting policy table by the presenter.
func (p PolicyViolation) GetViolation() string {
	for _, i := range p.Violation.ViolatedImages {
		return i.getViolationSubTable()
	}

	return ""
}

// Violations represents scanning violations made by images.
type Violations struct {
	ViolatedImages []Violation `json:"scanned"`
}

// Violation stores violation made by an image.
type Violation struct {
	Image           string          `json:"image"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

func (v Violation) getViolationSubTable() string {
	buf := bytes.Buffer{}
	defer buf.Reset()

	sortVulnerabilitiesBySeverities(v.Vulnerabilities)
	rows := make([][]string, 0, len(v.Vulnerabilities))

	for _, vulnerability := range v.Vulnerabilities {
		rows = append(rows, []string{
			strings.ToUpper(vulnerability.Severity),
			vulnerability.ID,
			vulnerability.Package,
			vulnerability.Type,
			vulnerability.FixAvailable,
		})
	}

	tabletool.GenerateTable(
		&buf,
		[]string{severityHeader, vulnerabilityHeader, packageHeader, typeHeader, fixAvailableHeader},
		rows,
		validatedImageOpts.Option)

	return strings.TrimSuffix(buf.String(), "\n")
}
