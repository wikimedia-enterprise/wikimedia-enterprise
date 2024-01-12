package schema

import "github.com/hamba/avro/v2"

// ConfigProbabilityScore schema configuration for ProbabilityScore.
var ConfigProbabilityScore = &Config{
	Type: ConfigTypeValue,
	Name: "ProbabilityScore",
	Schema: `{
		"type": "record",
		"name": "ProbabilityScore",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "prediction",
				"type": "boolean"
			},
			{
				"name": "probability",
				"type": [
					"null",
					{
						"name": "Probability",
						"type": "record",
						"fields": [
							{
								"name": "truthy",
								"type": "double"
							},
							{
								"name": "falsy",
								"type": "double"
							}
						]
					}
				]
			}
		]
	}`,
	Reflection: ProbabilityScore{},
}

// NewProbabilityScore creates new probability score avro schema.
func NewProbabilityScore() (avro.Schema, error) {
	return New(ConfigProbabilityScore)
}

// ProbabilityScore probability score representation for ORES models.
type ProbabilityScore struct {
	Prediction  bool         `json:"prediction,omitempty" avro:"prediction"`
	Probability *Probability `json:"probability,omitempty" avro:"probability"`
}
