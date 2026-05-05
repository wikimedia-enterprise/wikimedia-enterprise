package builder

import (
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
)

// LinkBuilder follows the builder pattern for Link.
type LinkBuilder struct {
	link *schema.Link
}

// NewLinkBuilder creates a Link builder.
func NewLinkBuilder() *LinkBuilder {
	return &LinkBuilder{
		link: new(schema.Link),
	}
}

// URL sets the URL field.
func (lb *LinkBuilder) URL(url string) *LinkBuilder {
	lb.link.URL = url
	return lb
}

// Text sets the Text field.
func (lb *LinkBuilder) Text(text string) *LinkBuilder {
	lb.link.Text = text
	return lb
}

// Images sets the Images field.
func (lb *LinkBuilder) Images(images []*schema.Image) *LinkBuilder {
	lb.link.Images = images
	return lb
}

// Build returns the constructed Link.
func (lb *LinkBuilder) Build() *schema.Link {
	return lb.link
}
