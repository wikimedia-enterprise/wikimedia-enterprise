package schema

import "github.com/hamba/avro/v2"

// ConfigArticleBody schema configuration for ArticleBody.
var ConfigArticleBody = &Config{
	Type: ConfigTypeValue,
	Name: "ArticleBody",
	Schema: `{
		"type": "record",
		"name": "ArticleBody",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "html",
				"type": "string"
			},
			{
				"name": "wikitext",
				"type": "string"
			}
		]
	}`,
	Reflection: ArticleBody{},
}

// NewArticleBodySchema creates new article body avro schema.
func NewArticleBodySchema() (avro.Schema, error) {
	return New(ConfigArticleBody)
}

// ArticleBody schema for article content.
// Not fully compliant with https://schema.org/articleBody, we need multiple article bodies.
type ArticleBody struct {
	HTML     string `json:"html,omitempty" avro:"html"`
	WikiText string `json:"wikitext,omitempty" avro:"wikitext"`
}
