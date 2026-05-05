package builder_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-contents/packages/builder"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type linkBuilderTestSuite struct {
	suite.Suite
	builder *builder.LinkBuilder
}

func (l *linkBuilderTestSuite) SetupTest() {
	l.builder = builder.NewLinkBuilder()
}

func (l *linkBuilderTestSuite) NewLinkBuilder() {
	lb := builder.NewLinkBuilder()
	l.Assert().NotNil(lb)

	lnk := lb.Build()
	l.Assert().NotNil(lnk)
}

func (l *linkBuilderTestSuite) TestUrl() {
	url := "en.wikicommons.org/a/b"
	lnk := l.builder.URL(url).Build()

	l.Assert().Equal(url, lnk.URL)
}

func (l *linkBuilderTestSuite) TestText() {
	txt := "...text.."
	lnk := l.builder.Text(txt).Build()

	l.Assert().Equal(txt, lnk.Text)
}

func (l *linkBuilderTestSuite) TestImages() {
	imgs := []*schema.Image{
		{ContentUrl: "commons.org/a"},
		{ContentUrl: "commons.org/b"},
	}
	lnk := l.builder.Images(imgs).Build()

	l.Assert().Equal(imgs, lnk.Images)
}

func TestLinkBuilder(t *testing.T) {
	suite.Run(t, new(linkBuilderTestSuite))
}
