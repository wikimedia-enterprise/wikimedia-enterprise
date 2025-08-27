package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type attributionTestSuite struct {
	suite.Suite
	attribution *Attribution
}

func (s *attributionTestSuite) SetupTest() {
	pfc := 25
	lrd := "06/25"

	editorsCount := &EditorsCount{
		AnonymousCount:  "<100",
		RegisteredCount: ">100",
	}

	s.attribution = &Attribution{
		EditorsCount:          editorsCount,
		ParsedReferencesCount: pfc,
		LastRevised:           lrd,
	}
}

func (s *attributionTestSuite) TestAttributionSchema() {
	sch, err := NewAttributionSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.attribution)
	s.Assert().NoError(err)

	attribution := new(Attribution)
	s.Assert().NoError(avro.Unmarshal(sch, data, attribution))
	s.Assert().Equal(attribution.EditorsCount.AnonymousCount, s.attribution.EditorsCount.AnonymousCount)
	s.Assert().Equal(attribution.EditorsCount.RegisteredCount, s.attribution.EditorsCount.RegisteredCount)
	s.Assert().Equal(attribution.ParsedReferencesCount, s.attribution.ParsedReferencesCount)
	s.Assert().Equal(attribution.LastRevised, s.attribution.LastRevised)
}

func (s *attributionTestSuite) TestAttributionSchemaWithNulls() {
	attribution := &Attribution{
		EditorsCount:          nil,
		ParsedReferencesCount: 0,
		LastRevised:           "",
	}

	sch, err := NewAttributionSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, attribution)
	s.Assert().NoError(err)

	result := new(Attribution)
	s.Assert().NoError(avro.Unmarshal(sch, data, result))
	s.Assert().Nil(result.EditorsCount)
	s.Assert().Equal(result.ParsedReferencesCount, 0)
	s.Assert().Equal(result.LastRevised, "")
}

func (s *attributionTestSuite) TestAttributionSchemaPartial() {
	pfc := 42

	attribution := &Attribution{
		EditorsCount:          nil,
		ParsedReferencesCount: pfc,
		LastRevised:           "",
	}

	sch, err := NewAttributionSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, attribution)
	s.Assert().NoError(err)

	result := new(Attribution)
	s.Assert().NoError(avro.Unmarshal(sch, data, result))
	s.Assert().Equal(result.ParsedReferencesCount, pfc)
	s.Assert().Equal(result.LastRevised, "")
}

func TestAttribution(t *testing.T) {
	suite.Run(t, new(attributionTestSuite))
}
