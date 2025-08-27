package schema

import (
	"github.com/hamba/avro/v2"
)

// ConfigStructuredTableRef schema configuration for StructuredTableRef.
var ConfigStructuredTableRef = &Config{
	Type: ConfigTypeValue,
	Name: "StructuredTableRef",
	Schema: `{
		"type": "record",
		"name": "StructuredTableRef",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "string",
				"default": ""
			},
			{
				"name": "confidence_score",
				"type": ["null", "double"],
				"default": null
			}
		]
	}`,
	References: []*Config{},
	Reflection: StructuredTableRef{},
}

// NewStructuredTablesRefSchema creates new StructuredTableRef avro schema.
func NewStructuredTablesRefSchema() (avro.Schema, error) {
	return New(ConfigStructuredTableRef)
}

type StructuredTableRef struct {
	Identifier      string   `json:"identifier,omitempty" avro:"identifier"`
	ConfidenceScore *float64 `json:"confidence_score,omitempty" avro:"confidence_score"`
}
