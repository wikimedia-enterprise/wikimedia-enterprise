package schema

import "github.com/hamba/avro/v2"

// ConfigEntity schema configuration for Entity.
var ConfigEntity = &Config{
	Type: ConfigTypeValue,
	Name: "Entity",
	Schema: `{
		"type": "record",
		"name": "Entity",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "string"
			},
			{
				"name": "url",
				"type": "string"
			},
			{
				"name": "aspects",
				"type": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			}
		]
	}`,
	Reflection: Entity{},
}

// NewEntitySchema creates new entity avro schema.
func NewEntitySchema() (avro.Schema, error) {
	return New(ConfigEntity)
}

// Entity schema for wikidata article.
// Right now will just be a copy of initial wikidata schema.
// Partially uses https://schema.org/Thing.
type Entity struct {
	Identifier string   `json:"identifier,omitempty" avro:"identifier"`
	URL        string   `json:"url,omitempty" avro:"url"`
	Aspects    []string `json:"aspects,omitempty" avro:"aspects"`
}
