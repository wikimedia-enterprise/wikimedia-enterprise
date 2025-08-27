package schema

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/hamba/avro/v2"
)

// ConfigStructured schema configuration for Structured.
var ConfigAvroStructured = &Config{
	Type: ConfigTypeValue,
	Name: "AvroStructured",
	Schema: `{
		"type": "record",
		"name": "AvroStructured",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "name",
				"type": "string"
			},
			{
				"name": "identifier",
				"type": "int"
			},
			{
				"name": "abstract",
				"type": "string",
				"default": ""
			},
			{
				"name": "description",
				"type": "string",
				"default": ""
			},
			{
				"name": "date_created",
				"type": [
					"null",
					{
						"type": "long",
						"logicalType": "timestamp-micros"
					}
				],
				"default": null
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
				"name": "version",
				"type": [
					"null",
					"Version"
				],
				"default": null
			},
			{
				"name": "url",
				"type": "string",
				"default": ""
			},
			{
				"name": "in_language",
				"type": [
					"null",
					"Language"
				],
				"default": null
			},
			{
				"name": "main_entity",
				"type": [
					"null",
					"Entity"
				],
				"default": null
			},
			{
				"name": "additional_entities",
				"type": {
					"type": "array",
					"items": {
						"type": "Entity"
					}
				}
			},
			{
				"name": "is_part_of",
				"type": [
					"null",
					"Project"
				],
				"default": null
			},
			{
				"name": "visibility",
				"type": [
					"null",
					"Visibility"
				],
				"default": null
			},
			{
				"name": "license",
				"type": {
					"type": "array",
					"items": {
						"type": "License"
					}
				}
			},
			{
				"name": "event",
				"type": [
					"null",
					"Event"
				],
				"default": null
			},
			{
				"name": "image",
				"type": [
					"null",
					"Image"
				],
				"default": null
			},
			{
				"name": "infoboxes",
                "type": {
					"type": "array",
					"items": {
						"type": "AvroHasParts"
					}
				}
				
			},
			{
				"name": "sections",
                "type": {
					"type": "array",
					"items": {
						"type": "AvroHasParts"
					}
				}
			},
			{
				"name": "references",
                "type": [ "null", 
					{
						"type": "array",
						"items": {
							"type": "StructuredReference"
						} 
					}
				],
				"default": null
			},
			{
				"name": "tables",
				"type": ["null",
					{
						"type": "array",
						"items": {
							"type": "AvroStructuredTable"
					}
					}
				],
				"default": null
			},
			{
				"name": "attribution",
				"type": [
					"null",
					"Attribution"
				],
				"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigAvroHasParts,
		ConfigEntity,
		ConfigVersion,
		ConfigLanguage,
		ConfigProject,
		ConfigLicense,
		ConfigEvent,
		ConfigImage,
		ConfigVisibility,
		ConfigStructuredReference,
		ConfigAvroStructuredTable,
		ConfigAttribution,
	},
	Reflection: AvroStructured{},
}

// NewStructuredSchema create new article avro schema.
func NewAvroStructuredSchema() (avro.Schema, error) {
	return New(ConfigAvroStructured)
}

// AvroStructured schema for wikipedia Avro Structured Contents.
type AvroStructured struct {
	TableName          struct{}               `json:"-" ksql:"structured"`
	Name               string                 `json:"name,omitempty" avro:"name"`
	Identifier         int                    `json:"identifier,omitempty" avro:"identifier"`
	Abstract           string                 `json:"abstract,omitempty" avro:"abstract"`
	Version            *Version               `json:"version,omitempty" avro:"version"`
	Event              *Event                 `json:"event,omitempty" avro:"event"`
	URL                string                 `json:"url,omitempty" avro:"url"`
	DateCreated        *time.Time             `json:"date_created,omitempty" avro:"date_created"`
	DateModified       *time.Time             `json:"date_modified,omitempty" avro:"date_modified"`
	MainEntity         *Entity                `json:"main_entity,omitempty" avro:"main_entity"`
	IsPartOf           *Project               `json:"is_part_of,omitempty" avro:"is_part_of"`
	AdditionalEntities []*Entity              `json:"additional_entities,omitempty" avro:"additional_entities"`
	InLanguage         *Language              `json:"in_language,omitempty" avro:"in_language"`
	Image              *Image                 `json:"image,omitempty" avro:"image"`
	Visibility         *Visibility            `json:"visibility,omitempty" avro:"visibility"`
	License            []*License             `json:"license,omitempty" avro:"license"`
	Description        string                 `json:"description,omitempty" avro:"description"`
	Infoboxes          []*AvroHasParts        `json:"infoboxes,omitempty" avro:"infoboxes"`
	Sections           []*AvroHasParts        `json:"sections,omitempty" avro:"sections"`
	References         []*StructuredReference `json:"references,omitempty" avro:"references"`
	Tables             []*AvroStructuredTable `json:"tables,omitempty" avro:"tables"`
	Attribution        *Attribution           `json:"attribution,omitempty" avro:"attribution"`
}

// Structured schema for wikipedia Structured Contents.
type Structured struct {
	TableName          struct{}               `json:"-" ksql:"structured"`
	Name               string                 `json:"name,omitempty" avro:"name"`
	Identifier         int                    `json:"identifier,omitempty" avro:"identifier"`
	Abstract           string                 `json:"abstract,omitempty" avro:"abstract"`
	Version            *Version               `json:"version,omitempty" avro:"version"`
	Event              *Event                 `json:"event,omitempty" avro:"event"`
	URL                string                 `json:"url,omitempty" avro:"url"`
	DateCreated        *time.Time             `json:"date_created,omitempty" avro:"date_created"`
	DateModified       *time.Time             `json:"date_modified,omitempty" avro:"date_modified"`
	MainEntity         *Entity                `json:"main_entity,omitempty" avro:"main_entity"`
	IsPartOf           *Project               `json:"is_part_of,omitempty" avro:"is_part_of"`
	AdditionalEntities []*Entity              `json:"additional_entities,omitempty" avro:"additional_entities"`
	InLanguage         *Language              `json:"in_language,omitempty" avro:"in_language"`
	Image              *Image                 `json:"image,omitempty" avro:"image"`
	Visibility         *Visibility            `json:"visibility,omitempty" avro:"visibility"`
	License            []*License             `json:"license,omitempty" avro:"license"`
	Description        string                 `json:"description,omitempty" avro:"description"`
	Infoboxes          []*HasParts            `json:"infoboxes,omitempty" avro:"infoboxes"`
	Sections           []*HasParts            `json:"sections,omitempty" avro:"sections"`
	References         []*StructuredReference `json:"references,omitempty" avro:"references"`
	Tables             []*AvroStructuredTable `json:"tables,omitempty" avro:"tables"`
	Attribution        *Attribution           `json:"attribution,omitempty" avro:"attribution"`
}

func (h *AvroStructured) ToJsonStruct() (*Structured, error) {
	data, err := json.Marshal(h)

	if err != nil {
		return nil, err
	}

	px := new(Structured)

	err = json.Unmarshal(data, px)

	if err != nil && !strings.Contains(err.Error(), "HasParts") {
		return nil, err
	}

	inf := make([]*HasParts, 0)

	for _, elem := range h.Infoboxes {
		avr, _ := elem.ToJsonStruct()
		inf = append(inf, avr)
	}

	px.Infoboxes = inf

	sec := make([]*HasParts, 0)

	for _, elem := range h.Sections {
		avr, _ := elem.ToJsonStruct()
		sec = append(sec, avr)
	}

	px.Sections = sec

	return px, nil
}

func (h *Structured) ToAvroStruct() (*AvroStructured, error) {
	data, err := json.Marshal(h)

	if err != nil {
		return nil, err
	}
	px := new(AvroStructured)

	err = json.Unmarshal(data, px)

	if err != nil && !strings.Contains(err.Error(), "HasParts") {
		return nil, err
	}

	inf := make([]*AvroHasParts, 0)

	for _, elem := range h.Infoboxes {
		avr, _ := elem.ToAvroStruct()
		inf = append(inf, avr)
	}

	px.Infoboxes = inf

	sec := make([]*AvroHasParts, 0)

	for _, elem := range h.Sections {
		avr, _ := elem.ToAvroStruct()
		sec = append(sec, avr)
	}

	px.Sections = sec

	return px, nil
}
