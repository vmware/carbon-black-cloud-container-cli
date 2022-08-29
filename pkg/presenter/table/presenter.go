// Package table provides utilities for showing results in table format
package table

import (
	"io"

	"github.com/vmware/carbon-black-cloud-container-cli/internal/util/tabletool"
)

// Option is the option used for table presenter.
type Option struct {
	// Limit is the number of rows to show in the result (table format only)
	Limit int
}

// Presenter will show the analysis result in table format.
type Presenter struct {
	provider Provider
	opts     Option
}

// NewPresenter will init a table presenter.
func NewPresenter(provider Provider, opts Option) *Presenter {
	return &Presenter{
		provider: provider,
		opts:     opts,
	}
}

// Title is the title of the table output.
func (p Presenter) Title() string {
	return p.provider.Title()
}

// Footer is the footer of the table output.
func (p Presenter) Footer() string {
	return p.provider.Footer()
}

// Present will convert the result into table format and pass to io.Writer.
func (p Presenter) Present(output io.Writer) error {
	rows := p.provider.Rows()
	tabletool.GenerateTable(output, p.provider.Header(), rows, tabletool.Option{
		Limit:       p.opts.Limit,
		WithBorder:  true,
		WithRowLine: true,
	})

	return nil
}
