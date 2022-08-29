// Package presenter provides utilities for showing results to the user in different format.
package presenter

import (
	"io"

	"github.com/vmware/carbon-black-cloud-container-cli/pkg/presenter/cyclondx"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/presenter/json"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/presenter/table"
)

// Presenter will show the analysis result to a given io.Writer.
type Presenter interface {
	Present(io.Writer) error
	Title() string
	Footer() string
}

// Provider is the base interface of present provider.
type Provider interface {
	Title() string
	Footer() string
}

// Option is the option used for presenter.
type Option struct {
	// OutputFormat is the output format of result format (table, json) of the report
	OutputFormat string
	// Limit is the number of rows to show in the result (table format only)
	Limit int
}

// NewPresenter will init a Presenter based on format.
func NewPresenter(provider Provider, opts Option) Presenter {
	switch opts.OutputFormat {
	case "json", "j":
		return json.NewPresenter(provider.(json.Provider))
	case "cyclonedx", "c":
		return cyclondx.NewPresenter(provider.(cyclondx.Provider))
	case "table", "t":
		fallthrough
	default:
		return table.NewPresenter(provider.(table.Provider), table.Option{Limit: opts.Limit})
	}
}
