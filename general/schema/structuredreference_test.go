package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type structuredReferenceTestSuite struct {
	suite.Suite
	ref *StructuredReference
}

func (s *structuredReferenceTestSuite) SetupTest() {

	metadata := map[string]string{}

	metadata["a"] = "b"

	referenceLink := &Link{
		URL:  "url",
		Text: "text",
		Images: []*Image{{
			ContentUrl: "https://localhost",
			Width:      100,
			Height:     100,
			Thumbnail: &Thumbnail{
				ContentUrl: "https://localhost",
				Width:      10,
				Height:     10,
			},
			AlternativeText: "altText",
			Caption:         "Caption",
		},
		},
	}

	referenceText := &StructuredReferenceText{
		Value: "value",
		Links: []*Link{referenceLink},
	}

	s.ref = &StructuredReference{
		Identifier: "id",
		Group:      "group",
		Type:       "type",
		Metadata:   metadata,
		Text:       referenceText,
		Source:     referenceText}

}

func (s *structuredReferenceTestSuite) TestStructuredReferenceSchema() {
	sch, err := NewStructuredReferenceSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.ref)
	s.Assert().NoError(err)

	reference := new(StructuredReference)
	s.Assert().NoError(avro.Unmarshal(sch, data, reference))

	s.Assert().Equal(reference.Identifier, s.ref.Identifier)
	s.Assert().Equal(reference.Group, s.ref.Group)
	s.Assert().Equal(reference.Type, s.ref.Type)
	s.Assert().Equal(reference.Metadata, s.ref.Metadata)
	s.Assert().Equal(reference.Text.Value, s.ref.Text.Value)
	s.Assert().Equal(reference.Source.Value, s.ref.Source.Value)

}

func TestStructuredReference(t *testing.T) {
	suite.Run(t, new(structuredReferenceTestSuite))
}
