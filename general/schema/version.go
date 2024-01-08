package schema

import "github.com/hamba/avro/v2"

// ConfigPreviousVersion schema configuration for PreviousVersion.
var ConfigPreviousVersion = &Config{
	Type: ConfigTypeValue,
	Name: "PreviousVersion",
	Schema: `{
		"type": "record",
		"name": "PreviousVersion",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "int"
			},
			{
				"name": "number_of_characters",
				"type": "int"
			}
		]
	}`,
}

// ConfigVersion schema configuration for Version.
var ConfigVersion = &Config{
	Type: ConfigTypeValue,
	Name: "Version",
	Schema: `{
		"type": "record",
		"name": "Version",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "identifier",
				"type": "int"
			},
			{
				"name": "comment",
				"type": "string"
			},
			{
				"name": "tags",
				"type": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			},
			{
				"name": "is_minor_edit",
				"type": "boolean"
			},
			{
				"name": "is_flagged_stable",
				"type": "boolean"
			},
			{
				"name": "has_tag_needs_citation",
				"type": "boolean",
				"default": false
			},
			{
				"name": "scores",
				"type": [
					"null",
					"Scores"
				]
			},
			{
				"name": "editor",
				"type": [
					"null",
					"Editor"
				]
			},
			{
				"name": "diff",
				"type": [
					"null",
					"Diff"
				],
				"default": null
			},
			{
				"name": "number_of_characters",
				"type": "int"
			},
			{
				"name": "sizes",
				"type": [
					"null",
					"Size"
				],
				"default": null
			},
			{
				"name": "is_breaking_news",
				"type": "boolean",
				"default": false
			},
			{
				"name": "noindex",
				"type": "boolean",
				"default": false
			},
			{
				"name": "maintenance_tags",
				"type": [
					"null",
					"MaintenanceTags"
				],
				"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigEvent,
		ConfigEditor,
		ConfigScores,
		ConfigSize,
		ConfigDiff,
		ConfigMaintenanceTags,
	},
	Reflection: Version{},
}

// NewPreviousVersionSchema creates a new previous version avro schema.
func NewPreviousVersionSchema() (avro.Schema, error) {
	return New(ConfigPreviousVersion)
}

// PreviousVersion is the representation for an article's previous version.
type PreviousVersion struct {
	Identifier         int `json:"identifier,omitempty" avro:"identifier"`
	NumberOfCharacters int `json:"number_of_characters,omitempty" avro:"number_of_characters"`
}

// NewVersionSchema creates new version avro schema.
func NewVersionSchema() (avro.Schema, error) {
	return New(ConfigVersion)
}

// Version representation for the article.
// Mainly modeled after https://schema.org/Thing.
type Version struct {
	TableName           struct{}         `json:"-" ksql:"versions"`
	Identifier          int              `json:"identifier,omitempty" avro:"identifier"`
	Comment             string           `json:"comment,omitempty" avro:"comment"`
	Tags                []string         `json:"tags,omitempty" avro:"tags"`
	IsMinorEdit         bool             `json:"is_minor_edit,omitempty" avro:"is_minor_edit"`
	IsFlaggedStable     bool             `json:"is_flagged_stable,omitempty" avro:"is_flagged_stable"`
	HasTagNeedsCitation bool             `json:"has_tag_needs_citation,omitempty" avro:"has_tag_needs_citation"`
	Scores              *Scores          `json:"scores,omitempty" avro:"scores"`
	Editor              *Editor          `json:"editor,omitempty" avro:"editor"`
	NumberOfCharacters  int              `json:"number_of_characters,omitempty" avro:"number_of_characters"`
	Size                *Size            `json:"size,omitempty" avro:"sizes"` // note that there's intentional `sizes` instead of `size` because size is ksqldb keyword
	Event               *Event           `json:"event,omitempty" avro:"event"`
	Diff                *Diff            `json:"diff,omitempty" avro:"diff"`
	IsBreakingNews      bool             `json:"is_breaking_news,omitempty" avro:"is_breaking_news"`
	Noindex             bool             `json:"noindex,omitempty" avro:"noindex"`
	MaintenanceTags     *MaintenanceTags `json:"maintenance_tags,omitempty" avro:"maintenance_tags"`
}
