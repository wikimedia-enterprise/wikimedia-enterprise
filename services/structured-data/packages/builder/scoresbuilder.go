package builder

import (
	"wikimedia-enterprise/services/structured-data/submodules/schema"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"
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

// ReferenceRisk sets the reference risk score in the builder's scores schema.
func (sb *ScoresBuilder) ReferenceRisk(sco *wmf.ReferenceRiskScore) *ScoresBuilder {
	if sco != nil {
		var survivalRatio *schema.SurvivalRatioData
		if sco.SurvivalRatio != nil {
			survivalRatio = &schema.SurvivalRatioData{
				Min:    sco.SurvivalRatio.Min,
				Mean:   sco.SurvivalRatio.Mean,
				Median: sco.SurvivalRatio.Median,
			}
		}
		// Ensure References is properly converted; if sco.References is nil, convertReferenceDetailsArray returns nil.
		sb.scores.ReferenceRisk = &schema.ReferenceRiskData{
			ReferenceCount:     sco.ReferenceCount,
			ReferenceRiskScore: sco.ReferenceRiskScore,
			SurvivalRatio:      survivalRatio,
			References:         convertReferenceDetailsArray(sco.References),
		}
	}
	return sb
}

// ReferenceNeed sets the reference need score in the builder's scores schema.
func (sb *ScoresBuilder) ReferenceNeed(sco *wmf.ReferenceNeedScore) *ScoresBuilder {
	if sco != nil {
		sb.scores.ReferenceNeed = &schema.ReferenceNeedData{
			ReferenceNeedScore: sco.ReferenceNeedScore,
		}
	}
	return sb
}

// Build returns the builder's scores schema.
func (sb *ScoresBuilder) Build() *schema.Scores {
	return sb.scores
}

// convertReferenceDetailsArray converts a slice of wmf.ReferenceDetails to a slice of schema.ReferenceDetails.
func convertReferenceDetailsArray(refs []*wmf.ReferenceDetails) []*schema.ReferenceDetails {
	if refs == nil {
		return nil
	}
	converted := make([]*schema.ReferenceDetails, len(refs))
	for i, ref := range refs {
		converted[i] = convertReferenceDetails(ref)
	}
	return converted
}

// convertReferenceDetails converts a single wmf.ReferenceDetails to a schema.ReferenceDetails.
func convertReferenceDetails(ref *wmf.ReferenceDetails) *schema.ReferenceDetails {
	if ref == nil {
		return nil
	}
	return &schema.ReferenceDetails{
		URL:            ref.URL,
		DomainName:     ref.DomainName,
		DomainMetadata: convertDomainMetadata(ref.DomainMetadata),
	}
}

// convertDomainMetadata converts a wmf.DomainMetadata to a schema.DomainMetadata.
func convertDomainMetadata(dm *wmf.DomainMetadata) *schema.DomainMetadata {
	if dm == nil {
		return nil
	}
	return &schema.DomainMetadata{
		PsLabelLocal:  dm.PsLabelLocal,
		PsLabelEnwiki: dm.PsLabelEnwiki,
		SurvivalRatio: dm.SurvivalRatio,
		PageCount:     dm.PageCount,
		EditorsCount:  dm.EditorsCount,
	}
}
