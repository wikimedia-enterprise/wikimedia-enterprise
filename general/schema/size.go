package schema

import "github.com/hamba/avro/v2"

// ConfigSize schema configuration for Size.
var ConfigSize = &Config{
	Type: ConfigTypeValue,
	Name: "Size",
	Schema: `{
		"type": "record",
		"name": "Size",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "value",
				"type": "double"
			},
			{
				"name": "unit_text",
				"type": "string"
			}
		]
	}`,
	Reflection: Size{},
}

// NewSizeSchema creates new size avro schema.
func NewSizeSchema() (avro.Schema, error) {
	return New(ConfigSize)
}

// Size representation according to https://schema.org/QuantitativeValue.
type Size struct {
	Value    float64 `json:"value,omitempty" avro:"value"`
	UnitText string  `json:"unit_text,omitempty" avro:"unit_text"`
}
