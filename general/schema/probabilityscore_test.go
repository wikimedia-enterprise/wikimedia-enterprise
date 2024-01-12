package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type probabilityScoreTestSuite struct {
	suite.Suite
	probabilityScore *ProbabilityScore
}

func (s *probabilityScoreTestSuite) SetupTest() {
	s.probabilityScore = &ProbabilityScore{
		Prediction: true,
		Probability: &Probability{
			True:  0.9,
			False: 0.1,
		},
	}
}

func (s *probabilityScoreTestSuite) TestNewProbabilityScoreSchema() {
	sch, err := NewProbabilityScore()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.probabilityScore)
	s.Assert().NoError(err)

	probabilityScore := new(ProbabilityScore)
	s.Assert().NoError(avro.Unmarshal(sch, data, probabilityScore))
	s.Assert().Equal(s.probabilityScore.Prediction, probabilityScore.Prediction)
	s.Assert().Equal(s.probabilityScore.Probability.True, probabilityScore.Probability.True)
	s.Assert().Equal(s.probabilityScore.Probability.False, probabilityScore.Probability.False)
}

func TestProbabilityScore(t *testing.T) {
	suite.Run(t, new(probabilityScoreTestSuite))
}
