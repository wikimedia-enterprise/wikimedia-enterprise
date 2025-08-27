package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type linkTestSuite struct {
	suite.Suite
	link *Link
}

func (s *linkTestSuite) SetupTest() {
	s.link = &Link{
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
}

func (s *linkTestSuite) TestNewLinkSchema() {
	sch, err := NewLinkSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.link)
	s.Assert().NoError(err)

	link := new(Link)
	s.Assert().NoError(avro.Unmarshal(sch, data, link))
	s.Assert().Equal(s.link.Images[0].ContentUrl, link.Images[0].ContentUrl)
	s.Assert().Equal(s.link.Images[0].Width, link.Images[0].Width)
	s.Assert().Equal(s.link.Images[0].Height, link.Images[0].Height)

	s.Assert().Equal(s.link.Images[0].Thumbnail.ContentUrl, link.Images[0].Thumbnail.ContentUrl)
	s.Assert().Equal(s.link.Images[0].Thumbnail.Width, link.Images[0].Thumbnail.Width)
	s.Assert().Equal(s.link.Images[0].Thumbnail.Height, link.Images[0].Thumbnail.Height)

	s.Assert().Equal(s.link.Images[0].AlternativeText, link.Images[0].AlternativeText)
	s.Assert().Equal(s.link.Images[0].Caption, link.Images[0].Caption)

}

func TestLink(t *testing.T) {
	suite.Run(t, new(linkTestSuite))
}
