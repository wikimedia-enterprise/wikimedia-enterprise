package schema

import (
	"github.com/hamba/avro/v2"
)

// ConfigHasParts schema configuration for StructuredReferenceText.
var ConfigStructuredReferenceText = &Config{
	Type: ConfigTypeValue,
	Name: "StructuredReferenceText",
	Schema: `{
		"type": "record",
		"name": "StructuredReferenceText",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
			"name": "value",
			"type": "string",
			"default": ""
			},
			{
			"name": "links",
			"type":{
					"type": "array",
					"items": {
					"type": "Link"
					}
				
				}
			}
		]
	}`,
	References: []*Config{
		ConfigLink,
	},
	Reflection: StructuredReferenceText{},
}

// NewStructuredReferenceTextSchema creates new StructuredReferenceText image avro schema.
func NewStructuredReferenceTextSchema() (avro.Schema, error) {
	return New(ConfigStructuredReferenceText)
}

type StructuredReferenceText struct {
	Value string  `json:"value,omitempty" avro:"value"`
	Links []*Link `json:"links,omitempty" avro:"links"`
}
