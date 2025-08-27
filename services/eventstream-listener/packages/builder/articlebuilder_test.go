package builder_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"wikimedia-enterprise/services/eventstream-listener/packages/builder"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

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
	s.builder = builder.NewArticleBuilder()
}

func (s *articleBuilderTestSuite) TestNewArticleBuilder() {
	ab := builder.NewArticleBuilder()
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

	article := s.builder.PreviousVersion(previousRevID).Build()
	s.Assert().Equal(previousRevID, article.PreviousVersion.Identifier)
}

func (s *articleBuilderTestSuite) TestVersionIdentifier() {
	versionIdentifier := "/enwiki/1"

	article := s.builder.VersionIdentifier(versionIdentifier).Build()
	s.Assert().Equal(versionIdentifier, article.VersionIdentifier)
}

func TestArticleBuilder(t *testing.T) {
	suite.Run(t, new(articleBuilderTestSuite))
}
