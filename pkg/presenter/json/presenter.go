// Package json provides utilities for showing results in json format
package json

import (
	"encoding/json"
	"io"
)

// Presenter will show the analyze result in json format.
type Presenter struct {
	provider Provider
}

// NewPresenter will init a JSONPresenter.
func NewPresenter(provider Provider) *Presenter {
	return &Presenter{
		provider: provider,
	}
}

// Title is the title of the json output.
func (p Presenter) Title() string {
	return p.provider.Title()
}

// Footer is the footer of the json output.
func (p Presenter) Footer() string {
	return p.provider.Footer()
}

// Present will convert the result into json format and pass to io.Writer.
func (p Presenter) Present(output io.Writer) error {
	enc := json.NewEncoder(output)
	// prevent > and < from being escaped in the payload
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")

	return enc.Encode(p.provider)
}
