package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type survivalRatioTestSuite struct {
	suite.Suite
	survivalRatio *SurvivalRatioData
}

func (s *survivalRatioTestSuite) SetupTest() {
	s.survivalRatio = &SurvivalRatioData{
		Min:    0.58,
		Mean:   0.65,
		Median: 0.60,
	}
}

func (s *survivalRatioTestSuite) TestNewSurvivalRatioSchema() {
	sch, err := NewSurvivalRatioData()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.survivalRatio)
	s.Assert().NoError(err)

	decoded := new(SurvivalRatioData)
	s.Assert().NoError(avro.Unmarshal(sch, data, decoded))
	s.Assert().Equal(s.survivalRatio.Min, decoded.Min)
	s.Assert().Equal(s.survivalRatio.Mean, decoded.Mean)
	s.Assert().Equal(s.survivalRatio.Median, decoded.Median)
}

func TestSurvivalRatio(t *testing.T) {
	suite.Run(t, new(survivalRatioTestSuite))
}
