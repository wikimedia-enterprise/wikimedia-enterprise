package builder

import (
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"
)

// StructuredReferencesBuilder follows the builder pattern for StructuredReferences.
type StructuredReferencesBuilder struct {
	ref *schema.StructuredReference
}

// NewHasPartsBuilder creates a new builder instance.
func NewStructuredReferenceBuilder() *StructuredReferencesBuilder {
	return &StructuredReferencesBuilder{
		ref: new(schema.StructuredReference),
	}
}

// FromPart initializes *schema.Hasparts using the values of *parser.Part
func (b *StructuredReferencesBuilder) FromParserReference(p *parser.Reference) *StructuredReferencesBuilder {

	b.ref.Identifier = p.Identifier
	b.ref.Metadata = p.Metadata
	b.ref.Group = p.Group
	b.ref.Type = p.Type

	if p.Text != nil {
		b.ref.Text = new(schema.StructuredReferenceText)

		b.ref.Text.Value = p.Text.Value
		b.ref.Text.Links = make([]*schema.Link, 0)

		for _, link := range p.Text.Links {
			lnk := NewLinkBuilder().Text(link.Text).URL(link.URL).Build()

			b.ref.Text.Links = append(b.ref.Text.Links, lnk)
		}
	}

	if p.Source != nil {
		b.ref.Source = new(schema.StructuredReferenceText)

		b.ref.Source.Value = p.Source.Value
		b.ref.Source.Links = make([]*schema.Link, 0)

		for _, link := range p.Source.Links {
			lnk := NewLinkBuilder().Text(link.Text).URL(link.URL).Build()

			b.ref.Source.Links = append(b.ref.Source.Links, lnk)
		}
	}
	return b
}

// Build returns the constructed HasParts.
func (b *StructuredReferencesBuilder) Build() *schema.StructuredReference {
	return b.ref
}
