package builder_test

import (
	"testing"
	"time"
	"wikimedia-enterprise/services/structured-contents/packages/builder"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type struncturedBuilderTestSuite struct {
	suite.Suite
	builder *builder.StructuredBuilder
}

// SetupTest initializes the structured builder.
func (s *struncturedBuilderTestSuite) SetupTest() {
	s.builder = builder.NewStructuredBuilder()
}

func (s *struncturedBuilderTestSuite) NewStructuredBuilder() {
	sb := builder.NewStructuredBuilder()
	s.Assert().NotNil(sb)

	str := sb.Build()
	s.Assert().NotNil(str)
}

func (s *struncturedBuilderTestSuite) TestEvent() {
	event := schema.NewEvent(schema.EventTypeCreate)
	str := s.builder.Event(event).Build()

	s.Assert().Equal(event, str.Event)
}

func (s *struncturedBuilderTestSuite) TestName() {
	name := "test name"
	str := s.builder.Name(name).Build()

	s.Assert().Equal(name, str.Name)
}

func (s *struncturedBuilderTestSuite) TestAbstract() {
	abstract := "...abstract goes here..."
	str := s.builder.Abstract(abstract).Build()

	s.Assert().Equal(abstract, str.Abstract)
}

func (s *struncturedBuilderTestSuite) TestIdentifier() {
	id := 1999999999
	str := s.builder.Identifier(id).Build()

	s.Assert().Equal(id, str.Identifier)
}

func (s *struncturedBuilderTestSuite) TestDateCreated() {
	date := time.Now()
	str := s.builder.DateCreated(&date).Build()

	s.Assert().Equal(&date, str.DateCreated)

	str = s.builder.DateCreated(nil).Build()
	s.Assert().Nil(str.DateCreated)
}

func (s *struncturedBuilderTestSuite) TestDateModified() {
	date := time.Now()
	str := s.builder.DateModified(&date).Build()

	s.Assert().Equal(&date, str.DateModified)

	str = s.builder.DateModified(nil).Build()
	s.Assert().Nil(str.DateModified)
}

func (s *struncturedBuilderTestSuite) TestURL() {
	wikiURL := "en.wikipedia.org"
	str := s.builder.URL(wikiURL).Build()

	s.Assert().Equal(wikiURL, str.URL)
}

func (s *struncturedBuilderTestSuite) TestInLanguage() {
	lang := &schema.Language{
		Identifier: "enwiki",
	}

	str := s.builder.InLanguage(lang).Build()

	s.Assert().Equal(lang, str.InLanguage)
}

func (s *struncturedBuilderTestSuite) TestIsPartOf() {
	project := &schema.Project{
		Identifier: "enwiki",
	}

	str := s.builder.IsPartOf(project).Build()

	s.Assert().Equal(project, str.IsPartOf)
}

func (s *struncturedBuilderTestSuite) TestLicense() {
	license := []*schema.License{schema.NewLicense()}

	str := s.builder.License(license).Build()
	s.Assert().Equal(str.License, license)
}

func (s *struncturedBuilderTestSuite) TestMainEntity() {
	baseItem := &schema.Entity{
		Identifier: "Q12",
		URL:        "wikidata.org/Q12",
	}

	str := s.builder.MainEntity(baseItem).Build()
	s.Assert().Equal(baseItem.Identifier, str.MainEntity.Identifier)
	s.Assert().Equal(baseItem.URL, str.MainEntity.URL)
}

func (s *struncturedBuilderTestSuite) TestAdditionalEntities() {
	additionalEntities := []*schema.Entity{
		{Aspects: []string{"example"}},
	}

	str := s.builder.AdditionalEntities(additionalEntities).Build()
	s.Assert().Equal(additionalEntities, str.AdditionalEntities)
}

func (s *struncturedBuilderTestSuite) TestVersion() {
	version := &schema.Version{
		Identifier: 166666666,
		Comment:    "Test comment",
	}

	str := s.builder.Version(version).Build()
	s.Assert().Equal(version, str.Version)
}

func (s *struncturedBuilderTestSuite) TestImage() {
	orgImg := &schema.Image{
		ContentUrl: "https://localhost/image.jpg",
		Width:      1000,
		Height:     1000,
	}

	str := s.builder.Image(orgImg).Build()

	s.Assert().NotNil(str.Image)
	s.Assert().Equal(orgImg.ContentUrl, str.Image.ContentUrl)
	s.Assert().Equal(orgImg.Width, str.Image.Width)
	s.Assert().Equal(orgImg.Height, str.Image.Height)
}

func (s *struncturedBuilderTestSuite) TesDescription() {
	des := "test "
	str := s.builder.Description(des).Build()

	s.Assert().Equal(des, str.Description)
}

func (s *struncturedBuilderTestSuite) TestInfoboxes() {
	inf := []*schema.AvroHasParts{
		{Name: "infobox"},
	}

	str := s.builder.Infoboxes(inf).Build()

	s.Assert().NotNil(str.Infoboxes)
	s.Assert().Equal(inf, str.Infoboxes)
}

func (s *struncturedBuilderTestSuite) TestSections() {
	sec := []*schema.AvroHasParts{
		{Name: "section 1"},
	}

	str := s.builder.Sections(sec).Build()

	s.Assert().NotNil(str.Sections)
	s.Assert().Equal(sec, str.Sections)
}

func (s *struncturedBuilderTestSuite) TestVisibility() {
	vis := &schema.Visibility{
		Text:    false,
		Editor:  false,
		Comment: false,
	}

	str := s.builder.Visibility(vis).Build()

	s.Assert().NotNil(str.Visibility)
	s.Assert().Equal(vis.Text, str.Visibility.Text)
	s.Assert().Equal(vis.Editor, str.Visibility.Editor)
	s.Assert().Equal(vis.Comment, str.Visibility.Comment)
}

func TestStructuredBuilder(t *testing.T) {
	suite.Run(t, new(struncturedBuilderTestSuite))
}
