package schema

import "github.com/hamba/avro/v2"

// ConfigProtection schema configuration for Protection.
var ConfigProtection = &Config{
	Type: ConfigTypeValue,
	Name: "Protection",
	Schema: `{
		"type": "record",
		"name": "Protection",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "type",
				"type": "string"
			},
			{
				"name": "level",
				"type": "string"
			},
			{
				"name": "expiry",
				"type": "string"
			}
		]
	}`,
	Reflection: Protection{},
}

// NewProtectionSchema creates new protection avro schema.
func NewProtectionSchema() (avro.Schema, error) {
	return New(ConfigProtection)
}

// Protection level for the article, does not comply with https://schema.org/ custom data.
type Protection struct {
	Type   string `json:"type,omitempty" avro:"type"`
	Level  string `json:"level,omitempty" avro:"level"`
	Expiry string `json:"expiry,omitempty" avro:"expiry"`
}
