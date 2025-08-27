package schema

import (
	"github.com/hamba/avro/v2"
)

// ConfigAttribution schema configuration for Attribution.
var ConfigAttribution = &Config{
	Type: ConfigTypeValue,
	Name: "Attribution",
	Schema: `{
		"type": "record",
		"name": "Attribution",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "editors_count",
				"type": ["null", "EditorsCount"],
				"default": null
			},
			{
				"name": "parsed_references_count",
				"type": "int",
				"default": 0
			},
			{
				"name": "last_revised",
				"type": "string",
				"default": ""
			}
		]
	}`,
	References: []*Config{
		ConfigEditorsCount,
	},
	Reflection: Attribution{},
}

// NewAttributionSchema creates new Attribution avro schema.
func NewAttributionSchema() (avro.Schema, error) {
	return New(ConfigAttribution)
}

// Attribution Schema for Attribution Trust Signals
type Attribution struct {
	EditorsCount          *EditorsCount `json:"editors_count,omitempty"  avro:"editors_count"`
	ParsedReferencesCount int           `json:"parsed_references_count,omitempty"  avro:"parsed_references_count"`
	LastRevised           string        `json:"last_revised,omitempty"  avro:"last_revised"`
}
