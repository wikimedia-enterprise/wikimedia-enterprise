package handler_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"
	"wikimedia-enterprise/general/parser"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articleupdate/handler"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/libraries/text"
	pb "wikimedia-enterprise/services/structured-data/packages/contentintegrity"

	"github.com/PuerkitoBio/goquery"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type unmarshalProducerMock struct {
	schema.UnmarshalProducer
	mock.Mock
}

func (m *unmarshalProducerMock) Unmarshal(_ context.Context, dta []byte, val interface{}) error {
	ags := m.Called(dta)

	switch val := val.(type) {
	case *schema.Key:
		*val = *ags.Get(0).(*schema.Key)
	case *schema.Article:
		*val = *ags.Get(0).(*schema.Article)
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

func (a *aggregatorMock) GetAggregation(_ context.Context, dtb string, ttl string, val *aggregate.Aggregation, _ ...aggregate.Getter) error {
	ags := a.Called(dtb, ttl)

	*val = *ags.Get(0).(*aggregate.Aggregation)

	return ags.Error(1)
}

type wordsPairGetterMock struct {
	mock.Mock
}

func (m *wordsPairGetterMock) GetWordsPair(_ context.Context, cts, prt []string) (*text.Words, *text.Words, error) {
	ags := m.Called(cts, prt)

	return ags.Get(0).(*text.Words), ags.Get(1).(*text.Words), ags.Error(2)
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

type integrityMock struct {
	mock.Mock
	pb.ContentIntegrityClient
}

func (i *integrityMock) GetArticleData(_ context.Context, _ *pb.ArticleDataRequest, _ ...grpc.CallOption) (*pb.ArticleDataResponse, error) {
	ags := i.Called()
	res := &pb.ArticleDataResponse{
		IsBreakingNews: ags.Get(0).(bool),
	}

	return res, ags.Error(1)
}

type protectedAPIMock struct {
	mock.Mock
	wmf.API
}

func (m *protectedAPIMock) GetPage(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*wmf.Page, error) {
	ags := m.Called(dtb, ttl)

	pge := ags.Get(0)
	if pge == nil {
		return nil, ags.Error(1)
	} else {
		return pge.(*wmf.Page), ags.Error(1)
	}
}

type TracerMock struct{}

func (t *TracerMock) Trace(ctx context.Context, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

func (t *TracerMock) StartTrace(ctx context.Context, _ string, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

func (t *TracerMock) Shutdown(ctx context.Context) error {
	return nil
}

type handlerTestSuite struct {
	suite.Suite
	ctx context.Context
	prs *handler.Parameters
	msg *kafka.Message
	key *schema.Key
	val *schema.Article
	agg *aggregate.Aggregation
	eku error
	evu error
	eag error
	ewp error
	epr error
	ebn error
	ibn bool
	bne bool
	msl int
	cts []string
	prt []string
	wds *text.Words
	abs string
	psc parser.Categories
	pst parser.Templates
}

func (s *handlerTestSuite) SetupSuite() {
	s.msg = &kafka.Message{
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	str := new(unmarshalProducerMock)
	str.On("Unmarshal", s.msg.Key).Return(s.key, s.eku)
	str.On("Unmarshal", s.msg.Value).Return(s.val, s.evu)
	str.On("Produce", s.msl).Return(s.epr)

	agg := new(aggregatorMock)
	agg.On("GetAggregation", s.val.IsPartOf.Identifier, s.val.Name).Return(s.agg, s.eag)

	wpg := new(wordsPairGetterMock)
	wpg.On("GetWordsPair", s.cts, s.prt).Return(s.wds, s.wds, s.ewp)

	prs := new(parserMock)
	prs.On("GetCategories").Return(s.psc)
	prs.On("GetTemplates").Return(s.pst)
	prs.On("GetAbstract").Return(s.abs)

	itg := new(integrityMock)
	itg.On("GetArticleData").Return(s.ibn, s.ebn)

	wmk := new(protectedAPIMock)

	rev := &wmf.Revision{Slots: &wmf.Slots{Main: &wmf.Main{Content: "BOYL"}}}
	wmk.On("GetPage", mock.Anything, mock.Anything, mock.Anything).Return(&wmf.Page{Revisions: []*wmf.Revision{rev}}, nil)

	s.ctx = context.Background()
	s.prs = &handler.Parameters{
		Env: &env.Environment{
			TopicArticles: "local.articles.v1",
			Topics: &schema.Topics{
				Versions: []string{"v1"},
			},
			BreakingNewsEnabled: s.bne,
		},
		Text:       wpg,
		Stream:     str,
		Aggregator: agg,
		Parser:     prs,
		Integrity:  itg,
		Tracer:     &TracerMock{},
	}
}

func (s *handlerTestSuite) TestArticleUpdate() {
	hdl := handler.NewArticleUpdate(s.prs)
	err := hdl(s.ctx, s.msg)

	if s.eku != nil {
		s.Assert().Equal(s.eku, err)
	} else if s.evu != nil {
		s.Assert().Equal(s.evu, err)
	} else if s.eag != nil {
		s.Assert().Equal(s.eag, err)
	} else if s.epr != nil {
		s.Assert().Equal(s.epr, err)
	} else if s.agg != nil && s.agg.GetPageHTMLError() != nil {
		s.Assert().Equal(s.agg.GetPageHTMLError(), err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*handlerTestSuite{
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeDelete),
			},
			eku: errors.New("key unmarshal error"),
			bne: false,
		},
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 10,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 0,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			eku: errors.New("key unmarshal error"),
			bne: true,
			ibn: true,
			ebn: nil,
		},
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 1,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			eku: errors.New("key unmarshal error"),
			bne: true,
			ebn: errors.New("grpc connection error to content integrity"),
		},
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			evu: errors.New("value unmarshal error"),
		},
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			agg: &aggregate.Aggregation{},
			eag: errors.New("aggregation query error"),
		},
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			agg: &aggregate.Aggregation{
				Page: &wmf.Page{
					Missing: true,
				},
			},
		},
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			agg: &aggregate.Aggregation{
				Page: &wmf.Page{},
				PageHTML: &wmf.PageHTML{
					Error: errors.New("html page missing"),
				},
			},
		},
		{
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			agg: &aggregate.Aggregation{
				Page: &wmf.Page{},
			},
			cts: []string{},
			prt: []string{},
			msl: 2,
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
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			agg: &aggregate.Aggregation{
				Page: &wmf.Page{
					Revisions: []*wmf.Revision{
						{},
						{},
					},
				},
			},
			cts: []string{},
			prt: []string{},
			wds: &text.Words{},
			msl: 2,
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
			key: &schema.Key{
				Type:       "update",
				Identifier: "/enwiki/Earth",
			},
			val: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
					InLanguage: &schema.Language{
						Identifier: "en",
					},
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				PreviousVersion: &schema.PreviousVersion{
					Identifier: 99,
				},
				Event: schema.NewEvent(schema.EventTypeUpdate),
			},
			agg: &aggregate.Aggregation{
				Page: &wmf.Page{
					Revisions: []*wmf.Revision{
						{},
						{},
					},
				},
			},
			cts: []string{},
			prt: []string{},
			wds: &text.Words{},
			msl: 2,
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
