package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type entityTestSuite struct {
	suite.Suite
	entity *Entity
}

func (s *entityTestSuite) SetupTest() {
	s.entity = &Entity{
		Identifier: "Q1",
		URL:        "http://test.com/Q1",
		Aspects:    []string{"A", "B", "C"},
	}
}

func (s *entityTestSuite) TestNewEntitySchema() {
	sch, err := NewEntitySchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.entity)
	s.Assert().NoError(err)

	entity := new(Entity)
	s.Assert().NoError(avro.Unmarshal(sch, data, entity))
	s.Assert().Equal(s.entity.Identifier, entity.Identifier)
	s.Assert().Equal(s.entity.URL, entity.URL)
	s.Assert().Equal(s.entity.Aspects, entity.Aspects)
}

func TestEntity(t *testing.T) {
	suite.Run(t, new(entityTestSuite))
}
