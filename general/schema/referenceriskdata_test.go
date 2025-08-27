package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type referenceRiskTestSuite struct {
	suite.Suite
	referenceRisk *ReferenceRiskData
}

func (s *referenceRiskTestSuite) SetupTest() {
	psLabelLocal := "reliable"
	psLabelEnwiki := "mostly_reliable"

	s.referenceRisk = &ReferenceRiskData{
		ReferenceCount:     5,
		ReferenceRiskScore: 0.75,
		SurvivalRatio: &SurvivalRatioData{
			Min:    0.5,
			Mean:   0.65,
			Median: 0.6,
		},
		References: []*ReferenceDetails{
			{
				URL:        "https://example.com",
				DomainName: "example.com",
				DomainMetadata: &DomainMetadata{
					PsLabelLocal:  &psLabelLocal,
					PsLabelEnwiki: &psLabelEnwiki,
					SurvivalRatio: 0.8,
					PageCount:     100,
					EditorsCount:  30,
				},
			},
		},
	}
}

func (s *referenceRiskTestSuite) TestNewReferenceRiskSchema() {
	sch, err := NewReferenceRiskData()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.referenceRisk)
	s.Assert().NoError(err)

	decoded := new(ReferenceRiskData)
	s.Assert().NoError(avro.Unmarshal(sch, data, decoded))

	s.Assert().Equal(s.referenceRisk.ReferenceCount, decoded.ReferenceCount)
	s.Assert().Equal(s.referenceRisk.ReferenceRiskScore, decoded.ReferenceRiskScore)
	s.Assert().Equal(s.referenceRisk.SurvivalRatio.Min, decoded.SurvivalRatio.Min)
	s.Assert().Equal(s.referenceRisk.SurvivalRatio.Mean, decoded.SurvivalRatio.Mean)
	s.Assert().Equal(s.referenceRisk.SurvivalRatio.Median, decoded.SurvivalRatio.Median)
	s.Assert().Equal(s.referenceRisk.References[0].URL, decoded.References[0].URL)
	s.Assert().Equal(s.referenceRisk.References[0].DomainName, decoded.References[0].DomainName)
}

func TestReferenceRisk(t *testing.T) {
	suite.Run(t, new(referenceRiskTestSuite))
}
