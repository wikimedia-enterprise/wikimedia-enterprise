package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type articleBodyTestSuite struct {
	suite.Suite
	articleBody *ArticleBody
}

func (s *articleBodyTestSuite) SetupTest() {
	s.articleBody = &ArticleBody{
		HTML:     "...article body goes here...",
		WikiText: "...article body wikitext goes here...",
	}
}

func (s *articleBodyTestSuite) TestNewArticleBodySchema() {
	sch, err := NewArticleBodySchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.articleBody)
	s.Assert().NoError(err)

	articleBody := new(ArticleBody)
	s.Assert().NoError(avro.Unmarshal(sch, data, articleBody))
	s.Assert().Equal(s.articleBody.HTML, articleBody.HTML)
	s.Assert().Equal(s.articleBody.WikiText, articleBody.WikiText)
}

func TestArticleBody(t *testing.T) {
	suite.Run(t, new(articleBodyTestSuite))
}
