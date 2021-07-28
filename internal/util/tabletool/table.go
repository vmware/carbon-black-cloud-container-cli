/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package tabletool

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// Option is the option used for table generation.
type Option struct {
	// Limit is the number of rows to show in the result (table format only)
	Limit int
	// WithBorder is whether to generate table with border
	WithBorder bool
	// WithRowLine is whether to generate table with row line
	WithRowLine bool
}

// GenerateTable is a tool to generate a *tablewriter.Table.
func GenerateTable(writer io.Writer, header []string, rows [][]string, opts Option) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)

	table.SetHeader(header)
	table.SetBorder(opts.WithBorder)
	table.SetRowLine(opts.WithRowLine)

	if opts.Limit > 0 && opts.Limit < len(rows) {
		table.AppendBulk(rows[:opts.Limit])

		omittedRow := make([]string, len(rows[0]))
		for i := range omittedRow {
			omittedRow[i] = "..."
		}

		table.Append(omittedRow)
	} else {
		table.AppendBulk(rows)
	}

	table.Render()

	return table
}
