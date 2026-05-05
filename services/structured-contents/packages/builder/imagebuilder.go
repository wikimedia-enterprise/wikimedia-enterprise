package builder

import (
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
)

// ImageBuilder follows the builder pattern for Image.
type ImageBuilder struct {
	image *schema.Image
}

// NewImageBuilder creates an Image builder.
func NewImageBuilder() *ImageBuilder {
	return &ImageBuilder{
		image: new(schema.Image),
	}
}

// ContentUrl sets the ContentUrl field.
func (ib *ImageBuilder) ContentUrl(url string) *ImageBuilder {
	ib.image.ContentUrl = url
	return ib
}

// AlternativeText sets the AlternativeText field.
func (ib *ImageBuilder) AlternativeText(text string) *ImageBuilder {
	ib.image.AlternativeText = text
	return ib
}

// Caption sets the Caption field.
func (ib *ImageBuilder) Caption(caption string) *ImageBuilder {
	ib.image.Caption = caption
	return ib
}

// Height sets the Height field.
func (ib *ImageBuilder) Height(height int) *ImageBuilder {
	ib.image.Height = height
	return ib
}

// Width sets the Width field.
func (ib *ImageBuilder) Width(width int) *ImageBuilder {
	ib.image.Width = width
	return ib
}

// Thumbnail sets the Thumbnail field.
func (ib *ImageBuilder) Thumbnail(thumbnail *schema.Thumbnail) *ImageBuilder {
	ib.image.Thumbnail = thumbnail
	return ib
}

// Build returns the constructed Image.
func (ib *ImageBuilder) Build() *schema.Image {
	return ib.image
}
