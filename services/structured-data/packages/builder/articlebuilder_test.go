package builder_test

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
	"wikimedia-enterprise/general/parser"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/packages/builder"

	"github.com/stretchr/testify/suite"
)

const (
	wikidataURL = "https://www.wikidata.org/entity/"
	wikiURL     = "https://en.wikipedia.org/"
)

type articleBuilderTestSuite struct {
	suite.Suite
	rdm     *rand.Rand
	builder *builder.ArticleBuilder
}

// SetupTest initializes the article builder and seeds a random number generator.
func (s *articleBuilderTestSuite) SetupTest() {
	s.rdm = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.builder = builder.NewArticleBuilder(wikiURL)
}

func (s *articleBuilderTestSuite) TestNewArticleBuilder() {
	ab := builder.NewArticleBuilder(wikiURL)
	s.Assert().NotNil(ab)

	article := ab.Build()
	s.Assert().NotNil(article)
}

func (s *articleBuilderTestSuite) TestEvent() {
	event := schema.NewEvent(schema.EventTypeCreate)
	article := s.builder.Event(event).Build()

	s.Assert().Equal(event, article.Event)
}

func (s *articleBuilderTestSuite) TestName() {
	name := "test name"
	article := s.builder.Name(name).Build()

	s.Assert().Equal(name, article.Name)
}

func (s *articleBuilderTestSuite) TestAbstract() {
	abstract := "...abstract goes here..."
	article := s.builder.Abstract(abstract).Build()

	s.Assert().Equal(abstract, article.Abstract)
}

func (s *articleBuilderTestSuite) TestWatchersCount() {
	watchersCount := 111
	article := s.builder.WatchersCount(watchersCount).Build()

	s.Assert().Equal(watchersCount, article.WatchersCount)
}
func (s *articleBuilderTestSuite) TestDatePreviouslyModified() {
	previouslyModified := time.Now()
	article := s.builder.DatePreviouslyModified(&previouslyModified).Build()

	s.Assert().Equal(&previouslyModified, article.DatePreviouslyModified)
}

func (s *articleBuilderTestSuite) TestIdentifier() {
	id := s.rdm.Int()
	article := s.builder.Identifier(id).Build()

	s.Assert().Equal(id, article.Identifier)
}

func (s *articleBuilderTestSuite) TestDateCreated() {
	date := time.Now()
	article := s.builder.DateCreated(&date).Build()

	s.Assert().Equal(&date, article.DateCreated)

	article = s.builder.DateCreated(nil).Build()
	s.Assert().Nil(article.DateCreated)
}

func (s *articleBuilderTestSuite) TestDateModified() {
	date := time.Now()
	article := s.builder.DateModified(&date).Build()

	s.Assert().Equal(&date, article.DateModified)

	article = s.builder.DateModified(nil).Build()
	s.Assert().Nil(article.DateModified)
}

func (s *articleBuilderTestSuite) TestURL() {
	article := s.builder.URL(wikiURL).Build()

	s.Assert().Equal(wikiURL, article.URL)
}

func (s *articleBuilderTestSuite) TestInLanguage() {
	lang := &schema.Language{
		Identifier: "enwiki",
	}
	article := s.builder.InLanguage(lang).Build()

	s.Assert().Equal(lang, article.InLanguage)
}

func (s *articleBuilderTestSuite) TestIsPartOf() {
	project := &schema.Project{
		Identifier: "enwiki",
	}
	article := s.builder.IsPartOf(project).Build()

	s.Assert().Equal(project, article.IsPartOf)
}

func (s *articleBuilderTestSuite) TestNamespace() {
	ns := &schema.Namespace{
		Identifier: schema.NamespaceArticle,
	}
	article := s.builder.Namespace(ns).Build()

	s.Assert().Equal(ns, article.Namespace)
}

func (s *articleBuilderTestSuite) TestBody() {
	wikiText := "some wiki text"
	html := "<!-- some html content --> "
	article := s.builder.Body(wikiText, html).Build()

	s.Assert().Equal(article.ArticleBody, &schema.ArticleBody{
		WikiText: wikiText,
		HTML:     html,
	})
}

func (s *articleBuilderTestSuite) TestLicense() {
	license := schema.NewLicense()

	article := s.builder.License(license).Build()
	s.Assert().Contains(article.License, license)
}

func (s *articleBuilderTestSuite) TestMainEntity() {
	baseItem := "Q2"

	article := s.builder.MainEntity(baseItem).Build()
	s.Assert().Equal(baseItem, article.MainEntity.Identifier)
	s.Assert().Equal(fmt.Sprintf("%s%s", wikidataURL, baseItem), article.MainEntity.URL)
}

func (s *articleBuilderTestSuite) TestAdditionalEntities() {
	additionalEntities := map[string]*wmf.WbEntityUsage{
		"Q2": {Aspects: []string{"example"}},
	}

	article := s.builder.AdditionalEntities(additionalEntities).Build()
	for id, entity := range additionalEntities {
		s.Assert().Contains(article.AdditionalEntities, &schema.Entity{
			Identifier: id,
			URL:        fmt.Sprintf("%s%s", wikidataURL, id),
			Aspects:    entity.Aspects,
		})
	}
}

func (s *articleBuilderTestSuite) TestCategories() {
	categories := []*parser.Category{
		{
			Name: "example",
			URL:  "http://localhost:8080/wiki/example",
		},
		{
			Name: "example 1",
			URL:  "http://localhost:8080/wiki/example 1",
		},
	}

	article := s.builder.Categories(categories).Build()

	for _, category := range categories {
		ctg := &schema.Category{
			Name: category.Name,
			URL:  category.URL,
		}

		s.Assert().Contains(article.Categories, ctg)
	}
}

func (s *articleBuilderTestSuite) TestProtection() {
	protection := []*wmf.Protection{
		{
			Type:   "example type",
			Level:  "example level",
			Expiry: "example expiry",
		},
	}

	article := s.builder.Protection(protection).Build()
	for _, p := range protection {
		s.Assert().Contains(article.Protection, &schema.Protection{
			Type:   p.Type,
			Level:  p.Level,
			Expiry: p.Expiry,
		})
	}
}

func (s *articleBuilderTestSuite) TestVersion() {
	version := &schema.Version{
		Identifier: s.rdm.Intn(100000),
		Comment:    "Test comment",
	}

	article := s.builder.Version(version).Build()
	s.Assert().Equal(version, article.Version)
}

func (s *articleBuilderTestSuite) TestPreviousVersion() {
	previousRevID := s.rdm.Intn(100000)
	previousRevContent := "some content"

	article := s.builder.PreviousVersion(previousRevID, previousRevContent).Build()
	s.Assert().Equal(previousRevID, article.PreviousVersion.Identifier)
	s.Assert().Equal(len([]rune(previousRevContent)), article.PreviousVersion.NumberOfCharacters)
}

func (s *articleBuilderTestSuite) TestTemplates() {
	templates := []*parser.Template{
		{
			Name: "example 1",
			URL:  "http://localhost:8080/example 1",
		},
		{
			Name: "example 2",
			URL:  "http://localhost:8080/example 2",
		},
	}

	article := s.builder.Templates(templates).Build()

	for _, template := range templates {
		s.Assert().Contains(article.Templates, &schema.Template{
			Name: template.Name,
			URL:  template.URL,
		})
	}
}

func (s *articleBuilderTestSuite) TestVersionIdentifier() {
	versionIdentifier := "/enwiki/1"

	article := s.builder.VersionIdentifier(versionIdentifier).Build()
	s.Assert().Equal(versionIdentifier, article.VersionIdentifier)
}

func (s *articleBuilderTestSuite) TestRedirects() {
	redirects := []*wmf.Redirect{
		{Title: "example 1"},
		{Title: "example 2"},
	}

	article := s.builder.Redirects(redirects).Build()
	for _, redirect := range redirects {
		s.Assert().Contains(article.Redirects, &schema.Redirect{
			Name: redirect.Title,
			URL:  fmt.Sprintf("%s/wiki/%s", wikiURL, strings.ReplaceAll(redirect.Title, " ", "_")),
		})
	}
}

func (s *articleBuilderTestSuite) TestImage() {
	orgImg := &wmf.Image{
		Source: "https://localhost/image.jpg",
		Width:  1000,
		Height: 1000,
	}
	thmbImg := &wmf.Image{
		Source: "https://localhost/thumb.jpg",
		Width:  100,
		Height: 100,
	}

	article := s.builder.Image(orgImg, thmbImg).Build()

	s.Assert().NotNil(article.Image)
	s.Assert().Equal(orgImg.Source, article.Image.ContentUrl)
	s.Assert().Equal(orgImg.Width, article.Image.Width)
	s.Assert().Equal(orgImg.Height, article.Image.Height)
	s.Assert().Equal(thmbImg.Source, article.Image.Thumbnail.ContentUrl)
	s.Assert().Equal(thmbImg.Width, article.Image.Thumbnail.Width)
	s.Assert().Equal(thmbImg.Height, article.Image.Thumbnail.Height)
}

func TestArticleBuilder(t *testing.T) {
	suite.Run(t, new(articleBuilderTestSuite))
}
