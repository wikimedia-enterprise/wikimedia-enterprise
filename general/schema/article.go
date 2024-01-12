package schema

import (
	"time"

	"github.com/hamba/avro/v2"
)

// ConfigArticle schema configuration for Article.
var ConfigArticle = &Config{
	Type: ConfigTypeValue,
	Name: "Article",
	Schema: `{
		"type": "record",
		"name": "Article",
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
				"name": "date_previously_modified",
				"type": [
					"null",
					{
						"type": "long",
						"logicalType": "timestamp-micros"
					}
				]
			},
			{
				"name": "protection",
				"type": {
					"type": "array",
					"items": {
						"type": "Protection"
					}
				}
			},
			{
				"name": "version",
				"type": [
					"null",
					"Version"
				]
			},
			{
				"name": "previous_version",
				"type": [
					"null",
					"PreviousVersion"
				]
			},
			{
        "name": "version_identifier",
        "type": "string"
      },
			{
				"name": "url",
				"type": "string"
			},
			{
				"name": "watchers_count",
				"type": "int",
				"default": 0
			},
			{
				"name": "namespace",
				"type": [
					"null",
					"Namespace"
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
				"name": "main_entity",
				"type": [
					"null",
					"Entity"
				]
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
				"name": "categories",
				"type": {
					"type": "array",
					"items": {
						"name": "Category",
						"type": "record",
						"fields": [
							{
								"name": "name",
								"type": "string"
							},
							{
								"name": "url",
								"type": "string"
							}
						]
					}
				}
			},
			{
				"name": "templates",
				"type": {
					"type": "array",
					"items": {
						"name": "Template",
						"type": "record",
						"fields": [
							{
								"name": "name",
								"type": "string"
							},
							{
								"name": "url",
								"type": "string"
							}
						]
					}
				}
			},
			{
				"name": "redirects",
				"type": {
					"type": "array",
					"items": {
						"name": "Redirect",
						"type": "record",
						"fields": [
							{
								"name": "name",
								"type": "string"
							},
							{
								"name": "url",
								"type": "string"
							}
						]
					}
				}
			},
			{
				"name": "is_part_of",
				"type": [
					"null",
					"Project"
				]
			},
			{
				"name": "article_body",
				"type": [
					"null",
					"ArticleBody"
				]
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
				"name": "visibility",
				"type": [
					"null",
					"Visibility"
				]
			},
			{
				"name": "event",
				"type": [
					"null",
					"Event"
				]
			},
			{
				"name": "image",
				"type": [
					"null",
					"Image"
				],
				"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigProtection,
		ConfigVersion,
		ConfigPreviousVersion,
		ConfigNamespace,
		ConfigLanguage,
		ConfigEntity,
		ConfigProject,
		ConfigArticleBody,
		ConfigLicense,
		ConfigVisibility,
		ConfigEvent,
		ConfigImage,
	},
	Reflection: Article{},
}

// NewArticleSchema create new article avro schema.
func NewArticleSchema() (avro.Schema, error) {
	return New(ConfigArticle)
}

// Article schema for wikipedia article.
// Tries to compliant with https://schema.org/Article.
type Article struct {
	TableName              struct{}         `json:"-" ksql:"articles"`
	Name                   string           `json:"name,omitempty" avro:"name"`
	Identifier             int              `json:"identifier,omitempty" avro:"identifier"`
	Abstract               string           `json:"abstract,omitempty" avro:"abstract"`
	DateCreated            *time.Time       `json:"date_created,omitempty" avro:"date_created"`
	DateModified           *time.Time       `json:"date_modified,omitempty" avro:"date_modified"`
	DatePreviouslyModified *time.Time       `json:"date_previously_modified,omitempty" avro:"date_previously_modified"`
	Protection             []*Protection    `json:"protection,omitempty" avro:"protection"`
	Version                *Version         `json:"version,omitempty" avro:"version"`
	PreviousVersion        *PreviousVersion `json:"previous_version,omitempty" avro:"previous_version"`
	VersionIdentifier      string           `json:"-" avro:"version_identifier"`
	URL                    string           `json:"url,omitempty" avro:"url"`
	WatchersCount          int              `json:"watchers_count,omitempty" avro:"watchers_count"`
	Namespace              *Namespace       `json:"namespace,omitempty" avro:"namespace"`
	InLanguage             *Language        `json:"in_language,omitempty" avro:"in_language"`
	MainEntity             *Entity          `json:"main_entity,omitempty" avro:"main_entity"`
	AdditionalEntities     []*Entity        `json:"additional_entities,omitempty" avro:"additional_entities"`
	Categories             []*Category      `json:"categories,omitempty" avro:"categories"`
	Templates              []*Template      `json:"templates,omitempty" avro:"templates"`
	Redirects              []*Redirect      `json:"redirects,omitempty" avro:"redirects"`
	IsPartOf               *Project         `json:"is_part_of,omitempty" avro:"is_part_of"`
	ArticleBody            *ArticleBody     `json:"article_body,omitempty" avro:"article_body"`
	License                []*License       `json:"license,omitempty" avro:"license"`
	Visibility             *Visibility      `json:"visibility,omitempty" avro:"visibility"`
	Event                  *Event           `json:"event,omitempty" avro:"event"`
	Image                  *Image           `json:"image,omitempty" avro:"image"`
}

// Category article category representation.
type Category struct {
	Name string `json:"name,omitempty" avro:"name"`
	URL  string `json:"url,omitempty" avro:"url"`
}

// Redirect article redirect representation.
type Redirect struct {
	Name string `json:"name,omitempty" avro:"name"`
	URL  string `json:"url,omitempty" avro:"url"`
}

// Template article template representation.
type Template struct {
	Name string `json:"name,omitempty" avro:"name"`
	URL  string `json:"url,omitempty" avro:"url"`
}
