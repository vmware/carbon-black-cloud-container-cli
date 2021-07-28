/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package colorizer provides utilities for showing results data.
package colorizer

import (
	"strings"

	"github.com/gookit/color"
)

const (
	// RiskHigh is a supported risk.
	RiskHigh = "HIGH"
	// RiskMedium is a supported risk.
	RiskMedium = "MEDIUM"
	// RiskLow is a supported risk.
	RiskLow = "LOW"
	// RiskUnknown is a supported risk.
	RiskUnknown = "UNKNOWN"
)

// ColorizeRisk will color the risk table item according the risk type.
func ColorizeRisk(risk string) string {
	switch strings.ToUpper(risk) {
	case RiskHigh:
		return color.Red.Sprint(RiskHigh)
	case RiskMedium:
		return color.LightRed.Sprint(RiskMedium)
	case RiskLow:
		return color.Yellow.Sprint(RiskLow)
	case RiskUnknown:
		fallthrough
	default:
		return RiskUnknown
	}
}
