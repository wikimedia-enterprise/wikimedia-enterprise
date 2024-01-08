package schema

import "github.com/hamba/avro/v2"

// ConfigImageObject schema configuration for ImageObject.
var ConfigThumbnail = &Config{
	Type: ConfigTypeValue,
	Name: "Thumbnail",
	Schema: `{
		"type": "record",
		"name": "Thumbnail",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "contentUrl",
				"type": "string"
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
	Reflection: Thumbnail{},
}

// NewThumbnailSchema creates new article image avro schema.
func NewThumbnailSchema() (avro.Schema, error) {
	return New(ConfigThumbnail)
}

// Thumbnail schema for article image thumbnail.
// Compliant with https://schema.org/ImageObject,
type Thumbnail struct {
	ContentUrl string `json:"content_url,omitempty" avro:"contentUrl"`
	Width      int    `json:"width,omitempty" avro:"width"`
	Height     int    `json:"height,omitempty" avro:"height"`
}
