package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type scoresTestSuite struct {
	suite.Suite
	scores *Scores
}

func (s *scoresTestSuite) SetupTest() {
	s.scores = &Scores{
		Damaging: &ProbabilityScore{
			Prediction: true,
			Probability: &Probability{
				True:  0.1,
				False: 0.9,
			},
		},
		GoodFaith: &ProbabilityScore{
			Prediction: true,
			Probability: &Probability{
				True:  0.1,
				False: 0.9,
			},
		},
		RevertRisk: &ProbabilityScore{
			Prediction: true,
			Probability: &Probability{
				True:  0.1,
				False: 0.9,
			},
		},
	}
}

func (s *scoresTestSuite) TestNewScoresSchema() {
	sch, err := NewScoresSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.scores)
	s.Assert().NoError(err)

	scores := new(Scores)
	s.Assert().NoError(avro.Unmarshal(sch, data, scores))
	s.Assert().Equal(s.scores.Damaging.Prediction, scores.Damaging.Prediction)
	s.Assert().Equal(s.scores.Damaging.Probability.False, scores.Damaging.Probability.False)
	s.Assert().Equal(s.scores.Damaging.Probability.True, scores.Damaging.Probability.True)
	s.Assert().Equal(s.scores.GoodFaith.Prediction, scores.GoodFaith.Prediction)
	s.Assert().Equal(s.scores.GoodFaith.Probability.False, scores.GoodFaith.Probability.False)
	s.Assert().Equal(s.scores.GoodFaith.Probability.True, scores.GoodFaith.Probability.True)
	s.Assert().Equal(s.scores.RevertRisk.Prediction, scores.RevertRisk.Prediction)
	s.Assert().Equal(s.scores.RevertRisk.Probability.False, scores.RevertRisk.Probability.False)
	s.Assert().Equal(s.scores.RevertRisk.Probability.True, scores.RevertRisk.Probability.True)
}

func TestScores(t *testing.T) {
	suite.Run(t, new(scoresTestSuite))
}
