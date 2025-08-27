package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type editorsCountTestSuite struct {
	suite.Suite
	editorsCount *EditorsCount
}

func (s *editorsCountTestSuite) SetupTest() {
	s.editorsCount = &EditorsCount{
		AnonymousCount:  "<100",
		RegisteredCount: ">100",
	}
}

func (s *editorsCountTestSuite) TestEditorsCountSchema() {
	sch, err := NewEditorsCountSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.editorsCount)
	s.Assert().NoError(err)

	editorsCount := new(EditorsCount)
	s.Assert().NoError(avro.Unmarshal(sch, data, editorsCount))

	s.Assert().Equal(editorsCount.AnonymousCount, s.editorsCount.AnonymousCount)
	s.Assert().Equal(editorsCount.RegisteredCount, s.editorsCount.RegisteredCount)
}

func TestEditorsCount(t *testing.T) {
	suite.Run(t, new(editorsCountTestSuite))
}
