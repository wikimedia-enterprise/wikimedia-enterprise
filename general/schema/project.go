package schema

import (
	"time"

	"github.com/hamba/avro/v2"
)

// ConfigProject schema configuration for Project.
var ConfigProject = &Config{
	Type: ConfigTypeValue,
	Name: "Project",
	Schema: `{
		"type": "record",
		"name": "Project",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "name",
				"type": "string"
			},
			{
				"name": "identifier",
				"type": "string"
			},
			{
				"name": "url",
				"type": "string"
			},
			{
				"name": "version",
				"type": "string"
			},
			{
				"name": "date_modified",
				"type": [
					"null",
					{
						"type": "long",
						"logicalType": "timestamp-micros"
					}
				]
			},
			{
				"name": "in_language",
				"type": [
					"null",
					"Language"
				]
			},
			{
				"name": "namespace",
				"type": [
					"null",
					"Namespace"
				],
				"default": null
			},
			{
				"name": "sizes",
				"type": [
					"null",
					"Size"
				]
			},
			{
				"name": "additional_type",
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
		ConfigLanguage,
		ConfigNamespace,
		ConfigSize,
	},
	Reflection: Project{},
}

// NewProjectSchema creates new project avro schema.
func NewProjectSchema() (avro.Schema, error) {
	return New(ConfigProject)
}

// Project representation of mediawiki project according to https://schema.org/Project.
type Project struct {
	Name           string     `json:"name,omitempty" avro:"name"`
	Identifier     string     `json:"identifier,omitempty" avro:"identifier"`
	URL            string     `json:"url,omitempty" avro:"url"`
	Version        string     `json:"version,omitempty" avro:"version"`
	AdditionalType string     `json:"additional_type,omitempty" avro:"additional_type"`
	Code           string     `json:"code,omitempty"`
	Namespace      *Namespace `json:"namespace,omitempty" avro:"namespace"`
	DateModified   *time.Time `json:"date_modified,omitempty" avro:"date_modified"`
	InLanguage     *Language  `json:"in_language,omitempty" avro:"in_language"`
	Size           *Size      `json:"size,omitempty" avro:"sizes"` // note that there's intentional `sizes` instead of `size` because size is ksqldb keyword
	Event          *Event     `json:"event,omitempty" avro:"event"`
}
