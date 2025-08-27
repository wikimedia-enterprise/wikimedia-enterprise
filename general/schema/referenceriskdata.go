package schema

import "github.com/hamba/avro/v2"

// ConfigReferenceRiskData schema configuration for ReferenceRiskScore.
var ConfigReferenceRiskData = &Config{
	Type: ConfigTypeValue,
	Name: "ReferenceRiskData",
	Schema: `{
		"namespace": "wikimedia_enterprise.general.schema",
		"name": "ReferenceRiskData",
		"type": "record",
		"fields": [
			{
				"name": "reference_count",
				"type": "int"
			},
			{
				"name": "reference_risk_score",
				"type": "double"
			},
			{
				"name": "survival_ratio",
				"type": [
					"null",
					"SurvivalRatioData"
				],
				"default": null
			},
			{
				"name": "references",
				"type": [
					"null",
					{
						"type": "array",
						"items": "ReferenceDetails"
					}
				],
				"default": null
			}
		]
	}`,
	References: []*Config{
		ConfigSurvivalRatioData,
		ConfigReferenceDetails,
	},
	Reflection: ReferenceRiskData{},
}

// NewReferenceRiskData creates new reference risk data avro schema.
func NewReferenceRiskData() (avro.Schema, error) {
	return New(ConfigReferenceRiskData)
}

// ReferenceRiskData represents the structured response data for the reference-need Liftwing API.
type ReferenceRiskData struct {
	ReferenceCount     int                 `json:"-" avro:"reference_count"` // Ignored in JSON, included in Avro
	ReferenceRiskScore float64             `json:"reference_risk_score" avro:"reference_risk_score"`
	SurvivalRatio      *SurvivalRatioData  `json:"-" avro:"survival_ratio"` //Ignored in JSON, included in Avro
	References         []*ReferenceDetails `json:"-" avro:"references"`     // Ignored in JSON, included in Avro
}
