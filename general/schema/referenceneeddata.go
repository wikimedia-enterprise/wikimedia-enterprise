package schema

import "github.com/hamba/avro/v2"

// ConfigReferenceNeedData schema configuration for ReferenceNeedScore.
var ConfigReferenceNeedData = &Config{
	Type: ConfigTypeValue,
	Name: "ReferenceNeedData",
	Schema: `{
		"type": "record",
		"name": "ReferenceNeedData",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "reference_need_score",
				"type": "double"
			}
		]
	}`,
	Reflection: ReferenceNeedData{},
}

// NewReferenceNeedData creates new reference need data avro schema.
func NewReferenceNeedData() (avro.Schema, error) {
	return New(ConfigReferenceNeedData)
}

// ReferenceNeedData represents the structured response data for the reference-need Liftwing API.
type ReferenceNeedData struct {
	ReferenceNeedScore float64 `json:"reference_need_score" avro:"reference_need_score"`
}
