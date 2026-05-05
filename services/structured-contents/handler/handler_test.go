package handler_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/structured-contents/config/env"
	"wikimedia-enterprise/services/structured-contents/handler"
	"wikimedia-enterprise/services/structured-contents/libraries/aggregate"
	"wikimedia-enterprise/services/structured-contents/submodules/config"
	pr "wikimedia-enterprise/services/structured-contents/submodules/prometheus"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"

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
	args := m.Called(dta)

	switch v := val.(type) {
	case *schema.Key:
		*v = *args.Get(0).(*schema.Key)
	case *schema.Article:
		*v = *args.Get(0).(*schema.Article)
	}

	return args.Error(1)
}

func (m *unmarshalProducerMock) Produce(_ context.Context, _ ...*schema.Message) error {
	return m.Called().Error(0)
}

type aggregatorMock struct {
	mock.Mock
	aggregate.Aggregator
}

func (a *aggregatorMock) GetAggregation(_ context.Context, dtb string, ttl string, doc *goquery.Document, val *aggregate.Aggregation, _ ...aggregate.Getter) error {
	ags := a.Called(dtb, ttl, doc)

	*val = *ags.Get(0).(*aggregate.Aggregation)

	return ags.Error(1)
}

type configMock struct {
	mock.Mock
	config.API
}

func (m *configMock) GetStructuredProjects() []string {
	return m.Called()[0].([]string)
}

type tracerMock struct {
	mock.Mock
}

func (t *tracerMock) Trace(ctx context.Context, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

func (t *tracerMock) StartTrace(ctx context.Context, _ string, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

func (t *tracerMock) Shutdown(ctx context.Context) error {
	return nil
}

type handlerTestSuite struct {
	suite.Suite
	ctx  context.Context
	prs  *handler.Parameters
	msg  *kafka.Message
	key  *schema.Key
	val  *schema.Article
	agg  *aggregate.Aggregation
	nct  string
	prjs []string
	eku  error
	evu  error
	eag  error
	epr  error
}

func (s *handlerTestSuite) TestHandler() {
	s.msg = &kafka.Message{
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	str := new(unmarshalProducerMock)
	str.On("Unmarshal", s.msg.Key).Return(s.key, s.eku)
	str.On("Unmarshal", s.msg.Value).Return(s.val, s.evu)
	str.On("Produce").Return(s.epr)

	agr := new(aggregatorMock)
	agr.On("GetAggregation", s.val.IsPartOf.Identifier, s.val.Name, mock.Anything).Return(s.agg, s.eag)

	trc := new(tracerMock)

	cfg := new(configMock)
	cfg.On("GetStructuredProjects").Return(s.prjs)

	s.ctx = context.Background()
	s.prs = &handler.Parameters{
		Env: &env.Environment{
			OutputCompactedTopics: &schema.Topics{
				Versions:    []string{"v1"},
				ServiceName: "structured-contents",
				Location:    "aws",
			},
			OutputNoncompactedTopic: s.nct,
		},
		Stream:     str,
		Aggregator: agr,
		Tracer:     trc,
		Cfg:        cfg,
		Metrics:    new(pr.Metrics),
	}

	hdl := handler.New(s.prs)
	err := hdl(s.ctx, s.msg)

	if s.eku != nil {
		s.Assert().Equal(s.eku, err)
	} else if s.evu != nil {
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
	for _, testCase := range []*handlerTestSuite{
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				Event: schema.NewEvent(schema.EventTypeDelete),
			},
			prjs: []string{"enwiki"},
		},
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				Event: schema.NewEvent(schema.EventTypeVisibilityChange),
			},
			prjs: []string{"enwiki"},
		},
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				ArticleBody: &schema.ArticleBody{
					HTML: "<html></html>",
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			agg: &aggregate.Aggregation{
				PageSummary: "desc",
				Infoboxes:   []*parser.Part{},
				Sections:    []*parser.Part{},
				References:  []*parser.Reference{},
			},
			prjs: []string{"enwiki"},
		},
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				ArticleBody: &schema.ArticleBody{
					HTML: "<html></html>",
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			nct: "aws.structured-contents.articles.v1",
			agg: &aggregate.Aggregation{
				PageSummary: "desc",
				Infoboxes:   []*parser.Part{},
				Sections:    []*parser.Part{},
				References:  []*parser.Reference{},
			},
			prjs: []string{"enwiki"},
		},
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			eku:  errors.New("key unmarshal error"),
			prjs: []string{"enwiki"},
		},
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			evu:  errors.New("val unmarshal error"),
			prjs: []string{"enwiki"},
		},
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				ArticleBody: &schema.ArticleBody{
					HTML: "<html></html>",
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			epr: errors.New("produce error"),
			agg: &aggregate.Aggregation{
				PageSummary: "desc",
				Infoboxes:   []*parser.Part{},
				Sections:    []*parser.Part{},
				References:  []*parser.Reference{},
			},
			prjs: []string{"enwiki"},
		},
		{
			key: &schema.Key{
				Type:       "article",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				ArticleBody: &schema.ArticleBody{
					HTML: "<html></html>",
				},
				Version: &schema.Version{
					Identifier: 100,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			prjs: []string{"enwiki"},
			nct:  "aws.structured-contents.articles.v1",
			eag:  errors.New("aggregation error"),
			agg: &aggregate.Aggregation{
				PageSummary: "desc",
				Infoboxes:   []*parser.Part{},
				Sections:    []*parser.Part{},
				References:  []*parser.Reference{},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
