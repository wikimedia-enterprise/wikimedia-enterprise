package schema

import "github.com/hamba/avro/v2"

// ConfigVisibility schema configuration for Visibility.
var ConfigVisibility = &Config{
	Type: ConfigTypeValue,
	Name: "Visibility",
	Schema: `{
		"type": "record",
		"name": "Visibility",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "text",
				"type": "boolean"
			},
			{
				"name": "editor",
				"type": "boolean"
			},
			{
				"name": "comment",
				"type": "boolean"
			}
		]
	}`,
	Reflection: Visibility{},
}

// NewVisibilitySchema create new avro schema for visibility.
func NewVisibilitySchema() (avro.Schema, error) {
	return New(ConfigVisibility)
}

// Visibility representing visibility changes for parts of the article.
// Custom dataset, not modeletd after https://schema.org/.
type Visibility struct {
	Text    bool `json:"text" avro:"text"`
	Editor  bool `json:"editor" avro:"editor"`
	Comment bool `json:"comment" avro:"comment"`
}
