package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type imageTestSuite struct {
	suite.Suite
	image *Image
}

func (s *imageTestSuite) SetupTest() {
	s.image = &Image{
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
	}
}

func (s *imageTestSuite) TestNewImageSchema() {
	sch, err := NewImageSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.image)
	s.Assert().NoError(err)

	image := new(Image)
	s.Assert().NoError(avro.Unmarshal(sch, data, image))
	s.Assert().Equal(s.image.ContentUrl, image.ContentUrl)
	s.Assert().Equal(s.image.Width, image.Width)
	s.Assert().Equal(s.image.Height, image.Height)

	s.Assert().Equal(s.image.Thumbnail.ContentUrl, image.Thumbnail.ContentUrl)
	s.Assert().Equal(s.image.Thumbnail.Width, image.Thumbnail.Width)
	s.Assert().Equal(s.image.Thumbnail.Height, image.Thumbnail.Height)

	s.Assert().Equal(s.image.AlternativeText, image.AlternativeText)
	s.Assert().Equal(s.image.Caption, image.Caption)

}

func TestImage(t *testing.T) {
	suite.Run(t, new(imageTestSuite))
}
