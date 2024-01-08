package handler_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articledelete/handler"

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

type apiMock struct {
	wmf.API
	mock.Mock
}

func (m *apiMock) GetPage(_ context.Context, dtb string, ttl string, _ ...func(*url.Values)) (*wmf.Page, error) {
	ags := m.Called(dtb, ttl)

	return ags.Get(0).(*wmf.Page), ags.Error(1)
}

type handlerTestSuite struct {
	suite.Suite
	ctx context.Context
	pms *handler.Parameters
	key *schema.Key
	art *schema.Article
	pge *wmf.Page
	msg *kafka.Message
	msl int
	euk error
	euv error
	egp error
	epc error
}

func (s *handlerTestSuite) SetupSuite() {
	pdm := new(unmarshalProducerMock)
	pdm.On("Unmarshal", s.msg.Key).Return(s.key, s.euk)
	pdm.On("Unmarshal", s.msg.Value).Return(s.art, s.euv)
	pdm.On("Produce", s.msl).Return(s.epc)

	apm := new(apiMock)
	apm.On("GetPage", s.art.IsPartOf.Identifier, s.art.Name).Return(s.pge, s.egp)

	s.ctx = context.Background()
	s.pms = &handler.Parameters{
		Stream: pdm,
		API:    apm,
		Env: &env.Environment{
			Topics: &schema.Topics{},
		},
	}
}

func (s *handlerTestSuite) TestNewArticleDelete() {
	hdl := handler.NewArticleDelete(s.pms)
	err := hdl(s.ctx, s.msg)

	if s.euk != nil {
		s.Assert().Equal(s.euk, err)
	} else if s.euv != nil {
		s.Assert().Equal(s.euv, err)
	} else if s.egp != nil && s.egp != wmf.ErrPageNotFound {
		s.Assert().Equal(s.egp, err)
	} else if s.epc != nil {
		s.Assert().Equal(s.epc, err)
	} else if s.pge != nil && s.pge.Title != s.art.Name {
		s.Assert().NoError(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*handlerTestSuite{
		{
			msg: &kafka.Message{
				Key: []byte("key"),
			},
			key: &schema.Key{},
			art: &schema.Article{
				Name: "Test",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{Identifier: 1},
			},
			euk: errors.New("key unmarshal failed"),
		},
		{
			msg: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
			},
			key: &schema.Key{},
			art: &schema.Article{
				Name: "Test",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{Identifier: 1},
			},
			euv: errors.New("value unmarshal failed"),
		},
		{
			msg: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
			},
			key: &schema.Key{},
			art: &schema.Article{
				Name: "Test",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{Identifier: 1},
				Event:   schema.NewEvent(schema.EventTypeDelete),
			},
			egp: errors.New("get page"),
		},
		{
			msl: 2,
			msg: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
			},
			key: &schema.Key{},
			art: &schema.Article{
				Name: "Test",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{Identifier: 1},
				Event:   schema.NewEvent(schema.EventTypeDelete),
			},
			egp: wmf.ErrPageNotFound,
		},
		{
			msl: 2,
			pge: &wmf.Page{
				Title:   "Test",
				Missing: true,
			},
			msg: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
			},
			key: &schema.Key{},
			art: &schema.Article{
				Name: "Test",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{Identifier: 1},
				Event:   schema.NewEvent(schema.EventTypeDelete),
			},
		},
		{
			msl: 2,
			pge: &wmf.Page{
				Title:   "Test",
				Missing: true,
			},
			msg: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
			},
			key: &schema.Key{},
			art: &schema.Article{
				Name: "Test",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{Identifier: 1},
				Event:   schema.NewEvent(schema.EventTypeDelete),
			},
			epc: errors.New("producer"),
		},
		{
			msl: 2,
			pge: &wmf.Page{
				Title:   "Marshall Mathers",
				Missing: false,
				Redirects: []*wmf.Redirect{
					{
						Title: "Eminem",
					},
				},
			},
			msg: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
			},
			key: &schema.Key{},
			art: &schema.Article{
				Name: "Eminem",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Namespace: &schema.Namespace{
					Identifier: 0,
				},
				Version: &schema.Version{
					Identifier: 1,
				},
				Event: schema.NewEvent(schema.EventTypeDelete),
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
