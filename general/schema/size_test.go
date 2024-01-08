package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type sizeTestSuite struct {
	suite.Suite
	size *Size
}

func (s *sizeTestSuite) SetupTest() {
	s.size = &Size{
		UnitText: "MB",
		Value:    50,
	}
}

func (s *sizeTestSuite) TestNewSizeSchema() {
	sch, err := NewSizeSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.size)
	s.Assert().NoError(err)

	size := new(Size)
	s.Assert().NoError(avro.Unmarshal(sch, data, size))
	s.Assert().Equal(s.size.UnitText, size.UnitText)
	s.Assert().Equal(s.size.Value, size.Value)
}

func TestSize(t *testing.T) {
	suite.Run(t, new(sizeTestSuite))
}
