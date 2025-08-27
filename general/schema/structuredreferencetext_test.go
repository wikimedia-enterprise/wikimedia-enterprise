package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type structuredReferenceTextTestSuite struct {
	suite.Suite
	refText *StructuredReferenceText
}

func (s *structuredReferenceTextTestSuite) SetupTest() {

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

	s.refText = &StructuredReferenceText{
		Value: "value",
		Links: []*Link{referenceLink},
	}
}

func (s *structuredReferenceTextTestSuite) TestStructuredReferenceSchema() {
	sch, err := NewStructuredReferenceTextSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.refText)
	s.Assert().NoError(err)

	reference := new(StructuredReferenceText)
	s.Assert().NoError(avro.Unmarshal(sch, data, reference))

	s.Assert().Equal(reference.Value, s.refText.Value)

	s.Assert().Equal(reference.Links[0].URL, s.refText.Links[0].URL)
	s.Assert().Equal(reference.Links[0].Text, s.refText.Links[0].Text)
	s.Assert().Equal(reference.Links[0].Images[0].ContentUrl, s.refText.Links[0].Images[0].ContentUrl)
	s.Assert().Equal(reference.Links[0].Images[0].Thumbnail, s.refText.Links[0].Images[0].Thumbnail)
	s.Assert().Equal(reference.Links[0].Images[0].Width, s.refText.Links[0].Images[0].Width)
	s.Assert().Equal(reference.Links[0].Images[0].Height, s.refText.Links[0].Images[0].Height)
	s.Assert().Equal(reference.Links[0].Images[0].AlternativeText, s.refText.Links[0].Images[0].AlternativeText)
	s.Assert().Equal(reference.Links[0].Images[0].Caption, s.refText.Links[0].Images[0].Caption)

}

func TestStructuredReferenceText(t *testing.T) {
	suite.Run(t, new(structuredReferenceTextTestSuite))
}
