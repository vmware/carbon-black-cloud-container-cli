// Package cyclondx provides utilities for showing results in cyclondx format
package cyclondx

import (
	"bytes"
	"encoding/xml"
	"io"
)

// Presenter will show the analyzed result in json format.
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
	doc := p.provider.CycloneDXDoc()
	encoder := xml.NewEncoder(output)
	r := bytes.NewReader(doc)

	_, err := r.WriteTo(output)
	if err != nil {
		return encoder.Encode(doc)
	}

	return encoder.Encode(output)
}
