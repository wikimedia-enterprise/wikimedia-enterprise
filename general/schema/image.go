package schema

import "github.com/hamba/avro/v2"

// ConfigImageObject schema configuration for ImageObject.
var ConfigImage = &Config{
	Type: ConfigTypeValue,
	Name: "Image",
	Schema: `{
		"type": "record",
		"name": "Image",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "contentUrl",
				"type": "string"
			},
			{
				"name": "thumbnail",
				"type": [
					"null",
					"Thumbnail"
				]
			},
			{
				"name": "width",
				"type": "int"
			},
			{
				"name": "height",
				"type": "int"
			}
		]
	}`,
	References: []*Config{
		ConfigThumbnail,
	},
	Reflection: Image{},
}

// NewImageSchema creates new article image avro schema.
func NewImageSchema() (avro.Schema, error) {
	return New(ConfigImage)
}

// Image schema for article image.
// Compliant with https://schema.org/ImageObject,
type Image struct {
	ContentUrl string     `json:"content_url,omitempty" avro:"contentUrl"`
	Thumbnail  *Thumbnail `json:"-" avro:"thumbnail"`
	Width      int        `json:"width,omitempty" avro:"width"`
	Height     int        `json:"height,omitempty" avro:"height"`
}
