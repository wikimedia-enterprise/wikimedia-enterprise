package schema

import "github.com/hamba/avro/v2"

// ConfigLink schema configuration for Link.
var ConfigLink = &Config{
	Type: ConfigTypeValue,
	Name: "Link",
	Schema: `{
		"type": "record",
		"name": "Link",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "url",
				"type": "string"
			},
			{
				"name": "text",
				"type": "string"
			},
			{
				"name": "images",
				"type": { 
							"type": "array",
							"items": { "type": "Image" }
						}
			}
			
		]
	}`,
	References: []*Config{
		ConfigImage,
	},
	Reflection: Link{},
}

// NewLinkSchema creates new article image avro schema.
func NewLinkSchema() (avro.Schema, error) {
	return New(ConfigLink)
}

// Link represents a link that can be found on a Wikipedia page.
type Link struct {
	URL    string   `json:"url,omitempty" avro:"url"`
	Text   string   `json:"text,omitempty" avro:"text"`
	Images []*Image `json:"images,omitempty" avro:"images"`
}
