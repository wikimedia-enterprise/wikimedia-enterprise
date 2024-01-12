package builder

import (
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
)

// ScoresBuilder follows the builder pattern for the Scores schema.
type ScoresBuilder struct {
	scores *schema.Scores
}

// NewScoresBuilder returns a scores builder with an empty scores schema.
func NewScoresBuilder() *ScoresBuilder {
	return &ScoresBuilder{new(schema.Scores)}
}

// RevertRisk sets the revert risk score in the builder's scores schema.
func (sb *ScoresBuilder) RevertRisk(sco *wmf.LiftWingScore) *ScoresBuilder {
	if sco != nil {
		sb.scores.RevertRisk = &schema.ProbabilityScore{
			Prediction: sco.Prediction,
			Probability: &schema.Probability{
				False: sco.Probability.False,
				True:  sco.Probability.True,
			},
		}
	}
	return sb
}

// Build returns the builder's scores schema.
func (sb *ScoresBuilder) Build() *schema.Scores {
	return sb.scores
}
