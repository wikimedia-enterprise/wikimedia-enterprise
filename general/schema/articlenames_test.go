package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type articleNamesTestSuite struct {
	suite.Suite
	articleNames *ArticleNames
}

func (s *articleNamesTestSuite) SetupTest() {
	dateCreated := time.Now().UTC()

	s.articleNames = &ArticleNames{
		Names: []string{"Earth", "Albert Einstein"},
		IsPartOf: &Project{
			Name:       "Wikipedia",
			Identifier: "enwiki",
		},
		Event: &Event{
			Identifier:  "ec369574-bf5e-4325-a720-c78d36a80cdb",
			Type:        "create",
			DateCreated: &dateCreated,
		},
	}
}

func (s *articleNamesTestSuite) TestNewArticleSchema() {
	sch, err := NewArticleNamesSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.articleNames)
	s.Assert().NoError(err)

	articleNames := new(ArticleNames)
	s.Assert().NoError(avro.Unmarshal(sch, data, articleNames))

	// article names fields
	s.Assert().Equal(s.articleNames.Names, articleNames.Names)
	s.Assert().Equal(s.articleNames.IsPartOf, articleNames.IsPartOf)

	// event fields
	s.Assert().Equal(s.articleNames.Event.Identifier, articleNames.Event.Identifier)
	s.Assert().Equal(s.articleNames.Event.Type, articleNames.Event.Type)
	s.Assert().Equal(s.articleNames.Event.DateCreated.Format(time.RFC1123), articleNames.Event.DateCreated.Format(time.RFC1123))
}

func TestArticleNamesSchema(t *testing.T) {
	suite.Run(t, new(articleNamesTestSuite))
}
