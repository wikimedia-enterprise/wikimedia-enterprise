package schema

import "github.com/hamba/avro/v2"

// ConfigArticleNames schema configuration for ArticleNames.
var ConfigArticleNames = &Config{
	Type: ConfigTypeValue,
	Name: "ArticleNames",
	Schema: `{
		"type": "record",
		"name": "ArticleNames",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "names",
				"type": { 
					"type": "array",
					"items": {
						"type": "string"
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
				"name": "event",
				"type": [
					"null",
					"Event"
				],
				"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigProject,
		ConfigEvent,
	},
	Reflection: ArticleNames{},
}

// NewConfigArticleNames creates new article names avro schema.
func NewArticleNamesSchema() (avro.Schema, error) {
	return New(ConfigArticleNames)
}

// ArticleNames schema for article names.
type ArticleNames struct {
	Names    []string `json:"names,omitempty" avro:"names"`
	Event    *Event   `json:"event,omitempty" avro:"event"`
	IsPartOf *Project `json:"is_part_of,omitempty" avro:"is_part_of"`
}
