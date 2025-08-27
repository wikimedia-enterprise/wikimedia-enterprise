package builder_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/packages/builder"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"

	"github.com/stretchr/testify/suite"
)

type scoresBuilderTestSuite struct {
	suite.Suite
	builder *builder.ScoresBuilder
}

func (s *scoresBuilderTestSuite) SetupTest() {
	s.builder = builder.NewScoresBuilder()
}

func (s *scoresBuilderTestSuite) TestNewScoresBuilder() {
	sb := builder.NewScoresBuilder()
	s.Assert().NotNil(sb)
}

func (s *scoresBuilderTestSuite) TestRevertRisk() {
	rvk := &wmf.LiftWingScore{
		Prediction: true,
		Probability: &wmf.BooleanProbability{
			True:  0.9,
			False: 0.1,
		},
	}

	scores := s.builder.RevertRisk(rvk).Build()
	s.Assert().Equal(scores.RevertRisk, &schema.ProbabilityScore{
		Prediction: rvk.Prediction,
		Probability: &schema.Probability{
			True:  rvk.Probability.True,
			False: rvk.Probability.False,
		},
	})
}

func TestScoresBuilderSuite(t *testing.T) {
	suite.Run(t, new(scoresBuilderTestSuite))
}
