package schema

import "github.com/hamba/avro/v2"

// ConfigSurvivalRatioData schema configuration for SurvivalRatio.
var ConfigSurvivalRatioData = &Config{
	Type: ConfigTypeValue,
	Name: "SurvivalRatioData",
	Schema: `{
		"type": "record",
		"name": "SurvivalRatioData",
		"namespace": "wikimedia_enterprise.general.schema",
		"fields": [
			{
			  "name": "min",
			  "type": "double"
			},
			{
			  "name": "mean",
			  "type": "double"
			},
			{
			  "name": "median",
			  "type": "double"
			}
		]
	}`,
	Reflection: SurvivalRatioData{},
}

// NewSurvivalRatioData creates new survival ratio data avro schema.
func NewSurvivalRatioData() (avro.Schema, error) {
	return New(ConfigSurvivalRatioData)
}

// SurvivalRatioData represents statistical survival ratio metrics,including the minimum, mean, and median values.
type SurvivalRatioData struct {
	Min    float64 `json:"min" avro:"min"`
	Mean   float64 `json:"mean" avro:"mean"`
	Median float64 `json:"median" avro:"median"`
}
