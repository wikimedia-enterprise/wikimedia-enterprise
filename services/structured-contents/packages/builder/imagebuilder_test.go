package builder_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-contents/packages/builder"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type imageBuilderTestSuite struct {
	suite.Suite
	builder *builder.ImageBuilder
}

func (i *imageBuilderTestSuite) SetupTest() {
	i.builder = builder.NewImageBuilder()
}

func (i *imageBuilderTestSuite) NewImageBuilder() {
	ib := builder.NewImageBuilder()
	i.Assert().NotNil(ib)

	img := ib.Build()
	i.Assert().NotNil(img)
}

func (i *imageBuilderTestSuite) TestContentUrl() {
	url := "en.wikicommons.org/a/b"
	img := i.builder.ContentUrl(url).Build()

	i.Assert().Equal(url, img.ContentUrl)
}

func (i *imageBuilderTestSuite) TestAlternateText() {
	txt := "...text.."
	img := i.builder.AlternativeText(txt).Build()

	i.Assert().Equal(txt, img.AlternativeText)
}

func (i *imageBuilderTestSuite) TestCaption() {
	txt := "...text.."
	img := i.builder.Caption(txt).Build()

	i.Assert().Equal(txt, img.Caption)
}

func (i *imageBuilderTestSuite) TestHeight() {
	hgt := 15
	img := i.builder.Height(hgt).Build()

	i.Assert().Equal(hgt, img.Height)
}

func (i *imageBuilderTestSuite) TestWidth() {
	wdt := 15
	img := i.builder.Width(wdt).Build()

	i.Assert().Equal(wdt, img.Width)
}

func (i *imageBuilderTestSuite) TestThumbnail() {
	tmb := &schema.Thumbnail{
		ContentUrl: "commons.org/foo",
	}
	img := i.builder.Thumbnail(tmb).Build()

	i.Assert().Equal(tmb, img.Thumbnail)
}

func TestImageBuilder(t *testing.T) {
	suite.Run(t, new(imageBuilderTestSuite))
}
