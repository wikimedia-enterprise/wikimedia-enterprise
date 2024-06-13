package articles_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"
	"wikimedia-enterprise/services/bulk-ingestion/handlers/articles"
	pb "wikimedia-enterprise/services/bulk-ingestion/handlers/protos"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type unmarshalProducerMock struct {
	schema.UnmarshalProducer
	mock.Mock
}

func (m *unmarshalProducerMock) Produce(_ context.Context, msgs ...*schema.Message) error {
	msg := msgs[0]
	msg.Value.(*schema.ArticleNames).Event = nil

	return m.Called(msg).Error(0)
}

type wmfMock struct {
	wmf.API
	mock.Mock
}

func (m *wmfMock) GetAllPages(_ context.Context, dtb string, cbk func([]*wmf.Page), ops ...func(*url.Values)) error {
	val := url.Values{}

	for _, opt := range ops {
		opt(&val)
	}

	ags := m.Called(dtb, val.Get("apnamespace"))
	pgs := ags.Get(0).([]*wmf.Page)
	cbk(pgs)

	return ags.Error(1)
}

func (m *wmfMock) GetProject(_ context.Context, dtb string) (*wmf.Project, error) {
	ags := m.Called(dtb)

	return ags.Get(0).(*wmf.Project), ags.Error(1)
}

type articlesTestSuite struct {
	suite.Suite
	ctx context.Context
	req *pb.ArticlesRequest
	pms *articles.Parameters
	prj *wmf.Project
	pgs []*wmf.Page
	nms []string
	aer error
	per error
}

func (s *articlesTestSuite) SetupSuite() {
	s.ctx = context.Background()

	msg := &schema.Message{
		Config: schema.ConfigArticleNames,
		Topic:  "articles",
		Key: &schema.Key{
			Identifier: fmt.Sprintf("article-names/%s/%d/%s", s.req.Project, s.req.Namespace, strings.Join(s.nms, "|")),
			Type:       schema.KeyTypeArticleNames,
		},
		Value: &schema.ArticleNames{
			Names: s.nms,
			IsPartOf: &schema.Project{
				Identifier: s.prj.DBName,
				URL:        s.prj.URL,
			},
		},
	}

	str := new(unmarshalProducerMock)
	str.On("Produce", msg).Return(nil)

	clt := new(wmfMock)
	clt.On("GetAllPages", s.req.Project, strconv.Itoa(int(s.req.Namespace))).Return(s.pgs, s.aer)
	clt.On("GetProject", s.req.Project).Return(s.prj, s.per)

	s.pms = &articles.Parameters{
		Stream: str,
		Client: clt,
		Env: &env.Environment{
			TopicArticles:    "articles",
			NumberOfArticles: 10,
		},
	}
}

func (s *articlesTestSuite) TestHandler() {
	res, err := articles.Handler(s.ctx, s.pms, s.req)

	if s.per != nil {
		s.Assert().Nil(res)
		s.Assert().Equal(s.per, err)
	} else if s.aer != nil {
		s.Assert().Nil(res)
		s.Assert().Equal(s.aer, err)
	} else {
		s.Assert().NotNil(res)
		s.Assert().Equal(int(res.Total), 1)
		s.Assert().Equal(int(res.Errors), 0)
	}
}

func TestArticles(t *testing.T) {
	req := &pb.ArticlesRequest{
		Project:   "enwiki",
		Namespace: 14,
	}
	nms := []string{
		"Earth",
		"Albert Einstein",
	}
	pgs := []*wmf.Page{
		{
			PageID: 1,
			Title:  "Earth",
		},
		{
			PageID: 2,
			Title:  "Albert Einstein",
		},
	}
	prj := &wmf.Project{
		URL:    "http://localhost",
		DBName: "eniwki",
	}

	for _, testCase := range []*articlesTestSuite{
		{
			req: req,
			pgs: pgs,
			nms: nms,
			prj: prj,
		},
		{
			req: req,
			pgs: pgs,
			nms: nms,
			prj: prj,
			aer: errors.New("error getting the project"),
		},
		{
			req: req,
			pgs: pgs,
			nms: nms,
			prj: prj,
			per: errors.New("error getting all pages"),
		},
	} {
		suite.Run(t, testCase)
	}
}
