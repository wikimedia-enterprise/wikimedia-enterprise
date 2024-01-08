package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/handlers/articlevisibility/handler"

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

type handlerTestSuite struct {
	suite.Suite
	ctx context.Context
	pms *handler.Parameters
	msg *kafka.Message
	art *schema.Article
	key *schema.Key
	msl int
	euk error
	euv error
	epc error
}

func (s *handlerTestSuite) SetupSuite() {
	pdm := new(unmarshalProducerMock)
	pdm.On("Unmarshal", s.msg.Key).Return(s.key, s.euk)
	pdm.On("Unmarshal", s.msg.Value).Return(s.art, s.euv)
	pdm.On("Produce", s.msl).Return(s.epc)

	s.ctx = context.Background()
	s.pms = &handler.Parameters{
		Stream: pdm,
		Env:    new(env.Environment),
	}
}

func (s *handlerTestSuite) TestNewArticleVisibility() {
	hdl := handler.NewArticleVisibility(s.pms)
	err := hdl(s.ctx, s.msg)

	if s.euk != nil {
		s.Assert().Equal(s.euk, err)
	} else if s.euv != nil {
		s.Assert().Equal(s.euv, err)
	} else if s.epc != nil {
		s.Assert().Equal(s.epc, err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*handlerTestSuite{
		{
			msg: &kafka.Message{
				Key: []byte("hey"),
			},
			key: new(schema.Key),
			euk: errors.New("key"),
		},
		{
			msg: &kafka.Message{
				Key:   []byte("hey"),
				Value: []byte("value"),
			},
			art: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				Event: schema.NewEvent(schema.EventTypeVisibilityChange),
			},
			key: new(schema.Key),
			euv: errors.New("value"),
		},
		{
			msl: 1,
			msg: &kafka.Message{
				Key:   []byte("hey"),
				Value: []byte("value"),
			},
			art: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				Event: schema.NewEvent(schema.EventTypeVisibilityChange),
			},
			key: new(schema.Key),
			epc: errors.New("produce"),
		},
		{
			msl: 1,
			msg: &kafka.Message{
				Key:   []byte("hey"),
				Value: []byte("value"),
			},
			art: &schema.Article{
				Name: "Earth",
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
					URL:        "http://localhost:8080",
				},
				InLanguage: &schema.Language{
					Identifier: "en",
				},
				Version: &schema.Version{
					Identifier: 100,
					Editor: &schema.Editor{
						DateStarted: &time.Time{},
						EditCount:   10,
						Groups:      []string{"group-1"},
					},
				},
				Event: schema.NewEvent(schema.EventTypeVisibilityChange),
			},
			key: new(schema.Key),
		},
	} {
		suite.Run(t, testCase)
	}
}
