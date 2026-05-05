package builder

import (
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"
)

// HasPartsBuilder follows the builder pattern for HasParts.
type HasPartsBuilder struct {
	parts *schema.HasParts
}

// NewHasPartsBuilder creates a new builder instance.
func NewHasPartsBuilder() *HasPartsBuilder {
	return &HasPartsBuilder{
		parts: new(schema.HasParts),
	}
}

// FromPart initializes *schema.Hasparts using the values of *parser.Part
func (b *HasPartsBuilder) FromPart(p *parser.Part) *HasPartsBuilder {
	b.parts.Name = p.Name
	b.parts.Type = p.Type
	b.parts.Value = p.Value
	b.parts.Values = p.Values

	// Convert Images using ImageBuilder
	for _, img := range p.Images {
		b.parts.Images = append(b.parts.Images, NewImageBuilder().
			ContentUrl(img.ContentUrl).
			AlternativeText(img.AlternativeText).
			Caption(img.Caption).
			Height(img.Height).
			Width(img.Width).
			Build())
	}

	// Convert Links using LinkBuilder
	for _, lnk := range p.Links {
		lib := NewLinkBuilder().URL(lnk.URL).Text(lnk.Text)
		lis := []*schema.Image{}

		for _, img := range lnk.Images {
			lis = append(lis, NewImageBuilder().
				ContentUrl(img.ContentUrl).
				AlternativeText(img.AlternativeText).
				Caption(img.Caption).
				Height(img.Height).
				Width(img.Width).
				Build())
		}

		b.parts.Links = append(b.parts.Links, lib.Images(lis).Build())
	}

	b.parts.Citations = make([]*schema.StructuredCitation, 0)
	for _, cit := range p.Citations {
		b.parts.Citations = append(b.parts.Citations, &schema.StructuredCitation{Identifier: cit.Identifier, Group: cit.Group, Text: cit.Text})
	}

	// Recursively convert HasParts
	for _, subPart := range p.HasParts {
		b.parts.HasParts = append(b.parts.HasParts, NewHasPartsBuilder().FromPart(subPart).Build())
	}

	return b
}

// Build returns the constructed HasParts.
func (b *HasPartsBuilder) Build() *schema.HasParts {
	return b.parts
}
