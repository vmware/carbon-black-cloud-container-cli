/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package image manages the image analysis subcommands.
package image

import (
	"github.com/spf13/cobra"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/presenter"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/scan"
)

type (
	scanOption      = scan.Option
	presenterOption = presenter.Option
)

var opts struct {
	scanOption
	presenterOption
}

// Cmd return the command related to image analysis.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Commands related to image analysis",
		Long:  `Commands related to image analysis`,
	}

	cmd.AddCommand(ScanCmd())
	cmd.AddCommand(ValidateCmd())

	cmd.PersistentFlags().StringVarP(
		&opts.OutputFormat, "output", "o", "table", "output format of the result")
	cmd.PersistentFlags().IntVar(
		&opts.Limit, "limit", 10,
		"number of rows to show in the report (for table format only; set to 0 will show all rows)")
	cmd.PersistentFlags().BoolVar(
		&opts.ShouldCleanup, "cleanup", false, "clean up image (for docker only) after scanning")
	cmd.PersistentFlags().BoolVar(
		&opts.ForceScan, "force", false, "trigger a force scan no matter the image is scanned or not")
	cmd.PersistentFlags().BoolVar(
		&opts.UseDockerDaemon, "use-docker", false, "use docker daemon to pull image")
	cmd.PersistentFlags().StringVar(
		&opts.Credential, "cred", "", "use `USERNAME[:PASSWORD]` for accessing the registry")
	cmd.PersistentFlags().IntVar(
		&opts.Timeout, "timeout", 600, "set the duration (second) for the scan process")

	return cmd
}
