package builder_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-contents/packages/builder"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"

	"github.com/stretchr/testify/suite"
)

type referencesBuilderTestSuite struct {
	suite.Suite
	p       *parser.Reference
	builder *builder.StructuredReferencesBuilder
}

// compareHasParts recursively compares all fields of parser.Part and schema.HasParts
func (s *referencesBuilderTestSuite) compareReferences(p *parser.Reference, ref *schema.StructuredReference) {
	s.Assert().Equal(p.Identifier, ref.Identifier)
	s.Assert().Equal(p.Group, ref.Group)
	s.Assert().Equal(p.Metadata, ref.Metadata)
	s.Assert().Equal(p.Type, ref.Type)
	s.Assert().Equal(p.Text.Value, ref.Text.Value)
}

func (s *referencesBuilderTestSuite) SetupTest() {
	s.builder = builder.NewStructuredReferenceBuilder()
}

func (s *referencesBuilderTestSuite) TestStructuredReferencesBuilder() {
	s.Assert().NotNil(s.builder)

	hpts := s.builder.Build()
	s.Assert().NotNil(hpts)
}

func (s *referencesBuilderTestSuite) TestFromPart() {
	hpts := s.builder.FromParserReference(s.p).Build()
	s.compareReferences(s.p, hpts)
}

func TestReferencesBuilder(t *testing.T) {
	for _, testCase := range []*referencesBuilderTestSuite{
		{
			p: &parser.Reference{
				Identifier: "id",
				Group:      "",
				Type:       "text",
				Metadata:   map[string]string{"a": "b"},
				Text:       &parser.RawData{Value: "value"},
				Source:     &parser.RawData{Value: "value"},
			},
		},
		{
			p: &parser.Reference{
				Identifier: "",
				Group:      "",
				Type:       "",
				Metadata:   map[string]string{},
				Text:       &parser.RawData{Value: ""},
				Source:     &parser.RawData{Value: ""},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
