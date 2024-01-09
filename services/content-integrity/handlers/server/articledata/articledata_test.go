package articledata_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/services/content-integrity/handlers/server/articledata"
	pb "wikimedia-enterprise/services/content-integrity/handlers/server/protos"
	"wikimedia-enterprise/services/content-integrity/libraries/integrity"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type integrityMock struct {
	integrity.API
	mock.Mock
}

func (m *integrityMock) GetArticle(_ context.Context, aps *integrity.ArticleParams, chs ...integrity.ArticleChecker) (*integrity.Article, error) {
	ags := m.Called(*aps, chs[0])

	if art := ags.Get(0); art != nil {
		return art.(*integrity.Article), ags.Error(1)
	}

	return nil, ags.Error(1)
}

type handlerTestSuite struct {
	suite.Suite
	ctx context.Context
	aps integrity.ArticleParams
	bns *integrity.BreakingNews
	req *pb.ArticleDataRequest
	igm *integrityMock
	hdr *articledata.Handler
	art *integrity.Article
	err error
}

func (s *handlerTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.err = errors.New("test")
	s.igm = new(integrityMock)
	s.bns = new(integrity.BreakingNews)
	s.hdr = &articledata.Handler{
		Integrity:    s.igm,
		BreakingNews: *s.bns,
	}
	tsn := timestamppb.Now()
	dtn := tsn.AsTime()

	s.req = &pb.ArticleDataRequest{
		Project:           "enwiki",
		Identifier:        100,
		VersionIdentifier: 1,
		Templates:         []string{"Template:Cite news"},
		Categories:        []string{"Category:Breaking news"},
		DateModified:      tsn,
	}
	s.aps = integrity.ArticleParams{
		Project:           s.req.GetProject(),
		Identifier:        int(s.req.GetIdentifier()),
		VersionIdentifier: int(s.req.GetVersionIdentifier()),
		Templates:         s.req.GetTemplates(),
		Categories:        s.req.GetCategories(),
		DateModified:      dtn,
	}

	tmn := time.Now()
	s.art = &integrity.Article{
		Project:            s.req.GetProject(),
		Identifier:         int(s.req.GetIdentifier()),
		VersionIdentifier:  int(s.req.GetVersionIdentifier()),
		EditsCount:         99,
		UniqueEditorsCount: 19,
		DateCreated:        &tmn,
		DateNamespaceMoved: &tmn,
		IsBreakingNews:     true,
	}
}

func (s *handlerTestSuite) TestGetArticleDataError() {
	s.igm.On("GetArticle", s.aps, s.bns).Return(nil, s.err)

	res, err := s.hdr.GetArticleData(s.ctx, s.req)
	s.Nil(res)
	s.Equal(s.err, err)
}

func (s *handlerTestSuite) TestGetArticleData() {
	s.igm.On("GetArticle", s.aps, s.bns).Return(s.art, nil)

	res, err := s.hdr.GetArticleData(s.ctx, s.req)
	s.NotNil(res)
	s.NoError(err)
	s.Equal(s.art.GetIdentifier(), int(res.GetIdentifier()))
	s.Equal(s.art.GetProject(), res.GetProject())
	s.Equal(s.art.GetVersionIdentifier(), int(res.GetVersionIdentifier()))
	s.Equal(s.art.GetEditsCount(), int(res.GetEditsCount()))
	s.Equal(s.art.GetUniqueEditorsCount(), int(res.GetUniqueEditorsCount()))
	s.Equal(s.art.GetDateCreated().Unix(), res.GetDateCreated().GetSeconds())
	s.Equal(s.art.GetDateNamespaceMoved().Unix(), res.GetDateNamespaceMoved().GetSeconds())
	s.Equal(s.art.GetIsBreakingNews(), res.GetIsBreakingNews())
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}
