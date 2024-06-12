package handler_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/general/parser"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articlebulk/handler"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"

	"github.com/PuerkitoBio/goquery"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type unmarshalProducerMock struct {
	schema.UnmarshalProducer
	mock.Mock
}

func (m *unmarshalProducerMock) Unmarshal(_ context.Context, dta []byte, val interface{}) error {
	ags := m.Called(dta)

	switch val := val.(type) {
	case *schema.ArticleNames:
		*val = *ags.Get(0).(*schema.ArticleNames)
	}

	return ags.Error(1)
}

func (m *unmarshalProducerMock) Produce(_ context.Context, mgs ...*schema.Message) error {
	return m.Called(len(mgs)).Error(0)
}

type aggregatorMock struct {
	mock.Mock
	aggregate.Aggregator
}

func (a *aggregatorMock) GetAggregations(_ context.Context, dtb string, tls []string, val map[string]*aggregate.Aggregation, _ ...aggregate.Getter) error {
	ags := a.Called(dtb, tls)

	for key, agr := range ags.Get(0).(map[string]*aggregate.Aggregation) {
		val[key] = agr
	}

	return ags.Error(1)
}

type parserMock struct {
	parser.API
	mock.Mock
}

func (m *parserMock) GetTemplates(_ *goquery.Selection) parser.Templates {
	ags := m.Called()

	return ags.Get(0).(parser.Templates)
}

func (m *parserMock) GetCategories(_ *goquery.Selection) parser.Categories {
	ags := m.Called()

	return ags.Get(0).(parser.Categories)
}

func (m *parserMock) GetAbstract(_ *goquery.Selection) (string, error) {
	return m.Called().String(0), nil
}

type handlerTestSuite struct {
	suite.Suite
	ctx  context.Context
	prs  *handler.Parameters
	msg  *kafka.Message
	val  *schema.ArticleNames
	aggs map[string]*aggregate.Aggregation
	evu  error
	eag  error
	epr  error
	msl  int
	abs  string
	psc  parser.Categories
	pst  parser.Templates
}

func (s *handlerTestSuite) SetupSuite() {
	s.msg = &kafka.Message{
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	str := new(unmarshalProducerMock)
	str.On("Unmarshal", s.msg.Value).Return(s.val, s.evu)
	str.On("Produce", s.msl).Return(s.epr)

	agg := new(aggregatorMock)
	agg.On("GetAggregations", s.val.IsPartOf.Identifier, s.val.Names).Return(s.aggs, s.eag)

	prs := new(parserMock)
	prs.On("GetCategories").Return(s.psc)
	prs.On("GetTemplates").Return(s.pst)
	prs.On("GetAbstract").Return(s.abs)

	s.ctx = context.Background()
	s.prs = &handler.Parameters{
		Env: &env.Environment{
			Topics: &schema.Topics{
				Versions: []string{"v1"},
			},
		},
		Stream:     str,
		Aggregator: agg,
		Parser:     prs,
	}
}

func (s *handlerTestSuite) TestArticleBulk() {
	hdl := handler.NewArticleBulk(s.prs)
	err := hdl(s.ctx, s.msg)

	if s.evu != nil {
		s.Assert().Equal(s.evu, err)
	} else if s.eag != nil {
		s.Assert().Equal(s.eag, err)
	} else if s.epr != nil {
		s.Assert().Equal(s.epr, err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestHandler(t *testing.T) {
	msg := &schema.ArticleNames{
		IsPartOf: &schema.Project{
			Identifier: "enwiki",
			URL:        "http://localhost:8080",
		},
		Names: []string{"Earth"},
		Event: schema.NewEvent(schema.EventTypeCreate),
	}

	for _, testCase := range []*handlerTestSuite{
		{
			val: msg,
		},
		{
			val: &schema.ArticleNames{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				Names: []string{"Earth"},
				Event: schema.NewEvent(schema.EventTypeCreate),
			},
			evu: errors.New("value unmarshal error"),
		},
		{
			val: &schema.ArticleNames{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				Names: []string{"Earth"},
				Event: schema.NewEvent(schema.EventTypeCreate),
			},
			aggs: map[string]*aggregate.Aggregation{},
			eag:  errors.New("aggregation query error"),
		},
		{
			val: &schema.ArticleNames{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				Names: []string{"Earth"},
				Event: schema.NewEvent(schema.EventTypeCreate),
			},
		},
		{
			val: &schema.ArticleNames{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				Names: []string{"Earth"},
				Event: schema.NewEvent(schema.EventTypeCreate),
			},
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						Missing: true,
					},
				},
			},
		},
		{
			val: &schema.ArticleNames{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				Names: []string{"Earth"},
				Event: schema.NewEvent(schema.EventTypeCreate),
			},
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{},
					PageHTML: &wmf.PageHTML{
						Error: errors.New("html page missing"),
					},
				},
			},
		},
		{
			val: &schema.ArticleNames{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				Names: []string{"Earth"},
				Event: schema.NewEvent(schema.EventTypeCreate),
			},
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{},
				},
			},
			msl: 1,
			psc: []*parser.Category{
				{
					Name: "Test",
					URL:  "http://localhost:8080/wiki/Test",
				},
			},
			pst: []*parser.Template{
				{
					Name: "Test",
					URL:  "http://localhost:8080/wiki/Test",
				},
			},
			abs: "...abstract goes here...",
		},
		{
			val: &schema.ArticleNames{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				Names: []string{"Earth"},
				Event: schema.NewEvent(schema.EventTypeCreate),
			},
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						Revisions: []*wmf.Revision{
							{},
							{},
						},
					},
				},
			},
			msl: 1,
			epr: errors.New("producer error"),
			psc: []*parser.Category{
				{
					Name: "Test",
					URL:  "http://localhost:8080/wiki/Test",
				},
			},
			pst: []*parser.Template{
				{
					Name: "Test",
					URL:  "http://localhost:8080/wiki/Test",
				},
			},
			abs: "...abstract goes here...",
		},
	} {
		suite.Run(t, testCase)
	}
}
