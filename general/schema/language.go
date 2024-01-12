package schema

import "github.com/hamba/avro/v2"

// ConfigLanguage schema configuration for Language.
var ConfigLanguage = &Config{
	Type: ConfigTypeValue,
	Name: "Language",
	Schema: `{
		"type": "record",
		"name": "Language",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "string"
			},
			{
				"name": "name",
				"type": "string"
			},
			{
				"name": "alternate_name",
				"type": "string"
			},
			{
				"name": "direction",
				"type": "string"
			},
			{
				"name": "event",
				"type": [
					"null",
					"Event"
				]
			}
		]
	}`,
	References: []*Config{
		ConfigEvent,
	},
	Reflection: Language{},
}

// NewLanguageSchema creates new language avro schema.
func NewLanguageSchema() (avro.Schema, error) {
	return New(ConfigLanguage)
}

// Language representation accroding to https://schema.org/Language.
type Language struct {
	Identifier    string `json:"identifier,omitempty" avro:"identifier"`
	Name          string `json:"name,omitempty" avro:"name"`
	AlternateName string `json:"alternate_name,omitempty" avro:"alternate_name"`
	Direction     string `json:"direction,omitempty" avro:"direction"`
	Event         *Event `json:"event,omitempty" avro:"event"`
}
