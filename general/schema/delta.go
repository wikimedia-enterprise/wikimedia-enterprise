package schema

import "github.com/hamba/avro/v2"

// ConfigDelta schema configuration for Delta.
var ConfigDelta = &Config{
	Type: ConfigTypeValue,
	Name: "Delta",
	Schema: `{
		"type": "record",
		"name": "Delta",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "increase",
				"type": "int"
			},
			{
				"name": "decrease",
				"type": "int"
			},
			{
				"name": "sum",
				"type": "int"
			},
			{
				"name": "proportional_increase",
				"type": "double"
			},
			{
				"name": "proportional_decrease",
				"type": "double"
			}
		]
	}`,
	Reflection: Delta{},
}

// NewDeltaSchema creates new delta avro schema.
func NewDeltaSchema() (avro.Schema, error) {
	return New(ConfigDelta)
}

// Delta represents the change description between two versions for certain dataset.
type Delta struct {
	Increase             int     `json:"increase" avro:"increase"`
	Decrease             int     `json:"decrease" avro:"decrease"`
	Sum                  int     `json:"sum" avro:"sum"`
	ProportionalIncrease float64 `json:"proportional_increase" avro:"proportional_increase"`
	ProportionalDecrease float64 `json:"proportional_decrease" avro:"proportional_decrease"`
}
