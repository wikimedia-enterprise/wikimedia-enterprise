package schema

import (
	"github.com/hamba/avro/v2"
)

// ConfigStructuredCitation schema configuration for StructuredCitation.
var ConfigStructuredCitation = &Config{
	Type: ConfigTypeValue,
	Name: "StructuredCitation",
	Schema: `{
		"type": "record",
		"name": "StructuredCitation",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
			"name": "identifier",
			"type": "string",
			"default": ""
			},
			{
			"name": "group",
			"type": "string",
			"default": ""
			},
			{
			"name": "text",
			"type": "string",
			"default": ""
			}
		]
	}`,
	References: []*Config{},
	Reflection: StructuredCitation{},
}

// NewStructuredCitationSchema creates new StructureCitationCitation image avro schema.
func NewStructuredCitationSchema() (avro.Schema, error) {
	return New(ConfigStructuredCitation)
}

type StructuredCitation struct {
	Identifier string `json:"identifier,omitempty" avro:"identifier"`
	Group      string `json:"group,omitempty" avro:"group"`
	Text       string `json:"text,omitempty" avro:"text"`
}
