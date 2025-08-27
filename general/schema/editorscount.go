package schema

import (
	"github.com/hamba/avro/v2"
)

// ConfigEditorsCount schema configuration for EditorsCount.
var ConfigEditorsCount = &Config{
	Type: ConfigTypeValue,
	Name: "EditorsCount",
	Schema: `{
		"type": "record",
		"name": "EditorsCount",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "anonymous_count",
				"type": "string",
				"default": ""
			},
			{
				"name": "registered_count",
				"type": "string",
				"default": ""
			}
		]
	}`,
	Reflection: EditorsCount{},
}

// NewEditorsCountSchema creates new EditorsCount avro schema.
func NewEditorsCountSchema() (avro.Schema, error) {
	return New(ConfigEditorsCount)
}

// EditorsCount represents the count of editors for an article relative to a threshold number.
type EditorsCount struct {
	AnonymousCount  string `json:"anonymous_count,omitempty"  avro:"anonymous_count"`
	RegisteredCount string `json:"registered_count,omitempty"  avro:"registered_count"`
}
