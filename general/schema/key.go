package schema

import "github.com/hamba/avro/v2"

// Available message keys.
const (
	KeyTypeArticle      = "articles"
	KeyTypeArticleNames = "articlenames"
	KeyTypeVersion      = "versions"
	KeyTypeProject      = "projects"
	KeyTypeLanguage     = "languages"
	KeyTypeNamespace    = "namespaces"
	KeyTypeStructured   = "structured"
)

// ConfigKey schema configuration for Key.
var ConfigKey = &Config{
	Type: ConfigTypeKey,
	Name: "Key",
	Schema: `{
		"type": "record",
		"name": "Key",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "string"
			},
			{
				"name": "type",
				"type": "string"
			}
		]
	}`,
	Reflection: Key{},
}

// NewKeySchema create new message key avro schema.
func NewKeySchema() (avro.Schema, error) {
	return New(ConfigKey)
}

// NewKey create new message key.
func NewKey(identifier string, keyType string) *Key {
	return &Key{
		Identifier: identifier,
		Type:       keyType,
	}
}

// Key schema for kafka message keys.
type Key struct {
	Identifier string `json:"identifier,omitempty" avro:"identifier"`
	Type       string `json:"type,omitempty" avro:"type"`
}
