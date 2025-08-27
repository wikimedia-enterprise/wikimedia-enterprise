package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type structuredCitationTestSuite struct {
	suite.Suite
	citation *StructuredCitation
}

func (s *structuredCitationTestSuite) SetupTest() {

	s.citation = &StructuredCitation{
		Identifier: "id",
		Group:      "group",
		Text:       "text"}

}

func (s *structuredCitationTestSuite) TestStructuredReferenceSchema() {
	sch, err := NewStructuredCitationSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.citation)
	s.Assert().NoError(err)

	citation := new(StructuredCitation)
	s.Assert().NoError(avro.Unmarshal(sch, data, citation))

	s.Assert().Equal(citation.Identifier, s.citation.Identifier)
	s.Assert().Equal(citation.Group, s.citation.Group)
	s.Assert().Equal(citation.Text, s.citation.Text)

}

func TestStructuredCitation(t *testing.T) {
	suite.Run(t, new(structuredCitationTestSuite))
}
