package handler_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articlebulk/handler"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/submodules/parser"
	"wikimedia-enterprise/services/structured-data/submodules/schema"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"

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

func (a *aggregatorMock) GetPageHTML(_ context.Context, dtb string, tls []string, _ int) (map[string]*aggregate.Aggregation, error) {
	ags := a.Called(dtb, tls)
	return ags.Get(0).(map[string]*aggregate.Aggregation), ags.Error(1)
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

type TracerMock struct{}

func (t *TracerMock) Trace(ctx context.Context, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

func (t *TracerMock) Shutdown(ctx context.Context) error {
	return nil
}

func (t *TracerMock) StartTrace(ctx context.Context, _ string, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
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
	agg.On("GetAggregations", s.val.IsPartOf.Identifier, []string{"5"}, mock.Anything, mock.Anything).Return(s.aggs, s.eag)
	agg.On("GetPageHTML").Return(s.aggs)

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
		Tracer:     &TracerMock{},
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

	var topic string = "topic"

	for _, testCase := range []*handlerTestSuite{
		{
			val: msg,
			msg: &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: 0,
					Offset:    0,
				},
				Key:   []byte("key"),
				Value: []byte("value"),
			},
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
					},
				},
			},
			msl: 1,
		},
		{
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
					},
				},
			},
			evu: errors.New("value unmarshal error"),
		},
		{
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
					},
				},
			},
			eag: errors.New("aggregation query error"),
		},
		{
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
					},
				},
			},
			msl: 1,
		},
		{
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						Missing:   true,
						LastRevID: 5,
						Title:     "Earth",
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
					},
				},
			},
		},
		{
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
					},
					PageHTML: &wmf.PageHTML{
						Error: errors.New("html page missing"),
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Error: errors.New("html page missing"),
					},
				},
			},
		},
		{
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
					},
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
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
						Revisions: []*wmf.Revision{
							{},
							{},
						},
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
						Title:   "Earth",
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
		{
			val: msg,
			aggs: map[string]*aggregate.Aggregation{
				"Earth": {
					Page: &wmf.Page{
						LastRevID: 5,
						Title:     "Earth",
						Revisions: []*wmf.Revision{
							{},
							{},
						},
					},
				},
				"5": {
					PageHTML: &wmf.PageHTML{
						Content: "<html></html>",
						Title:   "Earth",
					},
					Score: &wmf.Score{
						Output: &wmf.LiftWingScore{
							Prediction: true,
							Probability: &wmf.BooleanProbability{
								True:  1,
								False: 0,
							},
						},
					},
				},
			},
			msl: 1,
			epr: nil,
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
