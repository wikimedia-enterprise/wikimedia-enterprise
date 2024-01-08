package schema

import "github.com/hamba/avro/v2"

// Currently supported namespaces.
const (
	NamespaceArticle  = 0
	NamespaceFile     = 6
	NamespaceCategory = 14
	NamespaceTemplate = 10
)

// ConfigNamespace schema configuration for Namespace.
var ConfigNamespace = &Config{
	Type: ConfigTypeValue,
	Name: "Namespace",
	Schema: `{
		"type": "record",
		"name": "Namespace",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "name",
				"type": "string"
			},
			{
				"name": "alternate_name",
				"type": "string"
			},
			{
				"name": "identifier",
				"type": "int"
			},
			{
				"name": "description",
				"type": "string",
				"default": ""
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
		ConfigLanguage,
	},
	Reflection: Namespace{},
}

// NewNamespaceSchema creates new namespace avro schema.
func NewNamespaceSchema() (avro.Schema, error) {
	return New(ConfigNamespace)
}

// Namespace representation of mediawiki namespace.
// There's nothing related to this in https://schema.org/, we used  https://schema.org/Thing.
type Namespace struct {
	Name          string `json:"name,omitempty" avro:"name"`
	Identifier    int    `json:"identifier" avro:"identifier"`
	AlternateName string `json:"alternate_name,omitempty" avro:"alternate_name"`
	Description   string `json:"description,omitempty" avro:"description"`
	Event         *Event `json:"event,omitempty" avro:"event"`
}
