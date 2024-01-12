package schema

import "github.com/hamba/avro/v2"

// ConfigScores schema configuration for Scores.
var ConfigScores = &Config{
	Type: ConfigTypeValue,
	Name: "Scores",
	Schema: `{
		"type": "record",
		"name": "Scores",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
				"name": "damaging",
				"type": [
					"null",
					"ProbabilityScore"
				]
			},
			{
				"name": "goodfaith",
				"type": [
					"null",
					"ProbabilityScore"
				]
			},
			{
				"name": "revertrisk",
				"type": [
					"null",
					"ProbabilityScore"
				],
				"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigProbabilityScore,
	},
	Reflection: Scores{},
}

// NewScoresSchema creates new scores avro schema.
func NewScoresSchema() (avro.Schema, error) {
	return New(ConfigScores)
}

// Scores ORES scores representation, has nothing on https://schema.org/, it's a custom dataset.
// For more info https://ores.wikimedia.org/.
type Scores struct {
	Damaging   *ProbabilityScore `json:"damaging,omitempty" avro:"damaging"`
	GoodFaith  *ProbabilityScore `json:"goodfaith,omitempty" avro:"goodfaith"`
	RevertRisk *ProbabilityScore `json:"revertrisk,omitempty" avro:"revertrisk"`
}

// Probability numeric probability values form ORES models.
type Probability struct {
	False float64 `json:"false,omitempty" avro:"truthy"`
	True  float64 `json:"true,omitempty" avro:"falsy"`
}
