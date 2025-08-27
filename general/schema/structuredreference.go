package schema

import (
	"github.com/hamba/avro/v2"
)

// ConfigStructuredReference schema configuration for StructuredReference.
var ConfigStructuredReference = &Config{
	Type: ConfigTypeValue,
	Name: "StructuredReference",
	Schema: `{
		"type": "record",
		"name": "StructuredReference",
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
			"name": "type",
			"type": "string",
			"default": ""
			},
			{
			"name": "metadata",
			"type": {"type": "map", "values": "string"}
			},
			{
			"name": "text",
			"type": ["null", "StructuredReferenceText"],
			"default": null
			},
			{
			"name": "source",
			"type": ["null", "StructuredReferenceText"],
			"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigLink,
		ConfigStructuredReferenceText,
	},
	Reflection: StructuredReference{},
}

// NewStructuredReferenceSchema creates new StructuredReference image avro schema.
func NewStructuredReferenceSchema() (avro.Schema, error) {
	return New(ConfigStructuredReference)
}

type StructuredReference struct {
	Identifier string                   `json:"identifier,omitempty"  avro:"identifier"`
	Group      string                   `json:"group,omitempty"  avro:"group"`
	Type       string                   `json:"type,omitempty"  avro:"type"`
	Metadata   map[string]string        `json:"metadata,omitempty"  avro:"metadata"`
	Text       *StructuredReferenceText `json:"text,omitempty"  avro:"text"`
	Source     *StructuredReferenceText `json:"source,omitempty"  avro:"source"`
}
