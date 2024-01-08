package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type deltaTestSuite struct {
	suite.Suite
	delta *Delta
}

func (s *deltaTestSuite) SetupTest() {
	s.delta = &Delta{
		Increase:             10,
		Decrease:             5,
		Sum:                  100,
		ProportionalIncrease: 10,
		ProportionalDecrease: 5,
	}
}

func (s *deltaTestSuite) TestNewDeltaSchema() {
	sch, err := NewDeltaSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.delta)
	s.Assert().NoError(err)

	delta := new(Delta)
	s.Assert().NoError(avro.Unmarshal(sch, data, delta))
	s.Assert().Equal(s.delta.Increase, delta.Increase)
	s.Assert().Equal(s.delta.Decrease, delta.Decrease)
	s.Assert().Equal(s.delta.Sum, delta.Sum)
	s.Assert().Equal(s.delta.ProportionalIncrease, delta.ProportionalIncrease)
	s.Assert().Equal(s.delta.ProportionalDecrease, delta.ProportionalDecrease)
}

func TestDelta(t *testing.T) {
	suite.Run(t, new(deltaTestSuite))
}
