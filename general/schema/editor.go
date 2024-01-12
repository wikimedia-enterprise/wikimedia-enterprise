package schema

import (
	"time"

	"github.com/hamba/avro/v2"
)

// ConfigEditor schema configuration for Editor.
var ConfigEditor = &Config{
	Type: ConfigTypeValue,
	Name: "Editor",
	Schema: `{
		"type": "record",
		"name": "Editor",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "int"
			},
			{
				"name": "name",
				"type": "string"
			},
			{
				"name": "edit_count",
				"type": "int"
			},
			{
				"name": "groups",
				"type": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			},
			{
				"name": "is_bot",
				"type": "boolean"
			},
			{
				"name": "is_anonymous",
				"type": "boolean"
			},
			{
				"name": "is_admin",
				"type": "boolean",
				"default": false
			},
			{
				"name": "is_patroller",
				"type": "boolean",
				"default": false
			},
			{
				"name": "has_advanced_rights",
				"type": "boolean",
				"default": false
			},
			{
				"name": "date_started",
				"type": [
					"null",
					{
						"type": "long",
						"logicalType": "timestamp-micros"
					}
				]
			}
		]
	}`,
	Reflection: Editor{},
}

// NewEditorSchema creates new editor avro schema.
func NewEditorSchema() (avro.Schema, error) {
	return New(ConfigEditor)
}

// Editor for the article version.
// Combines Person and CreativeWork with custom properties, link https://schema.org/editor.
type Editor struct {
	Identifier        int        `json:"identifier,omitempty" avro:"identifier"`
	Name              string     `json:"name,omitempty" avro:"name"`
	EditCount         int        `json:"edit_count,omitempty" avro:"edit_count"`
	Groups            []string   `json:"groups,omitempty" avro:"groups"`
	IsBot             bool       `json:"is_bot,omitempty" avro:"is_bot"`
	IsAnonymous       bool       `json:"is_anonymous,omitempty" avro:"is_anonymous"`
	IsAdmin           bool       `json:"is_admin,omitempty" avro:"is_admin"`
	IsPatroller       bool       `json:"is_patroller,omitempty" avro:"is_patroller"`
	HasAdvancedRights bool       `json:"has_advanced_rights,omitempty" avro:"has_advanced_rights"`
	DateStarted       *time.Time `json:"date_started,omitempty" avro:"date_started"`
}
