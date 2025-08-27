package handler_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/handlers/articles/handler"
	"wikimedia-enterprise/services/on-demand/submodules/prometheus"
	"wikimedia-enterprise/services/on-demand/submodules/schema"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type storageMock struct {
	mock.Mock
}

func (m *storageMock) Update(_ context.Context, kdt []byte, etp string, v interface{}) error {
	return m.Called(kdt, etp, v).Error(0)
}

type unmarshalerMock struct {
	mock.Mock
	schema.UnmarshalProducer
}

func (m *unmarshalerMock) Unmarshal(_ context.Context, data []byte, v interface{}) error {
	ags := m.Called(data)

	*v.(*schema.Article) = *ags.Get(0).(*schema.Article)

	return ags.Error(1)
}

type handlerTestSuite struct {
	suite.Suite
	ctx context.Context
	stg *storageMock
	stm *unmarshalerMock
	prs *handler.Parameters
	msg *kafka.Message
	sme error
	ste error
}

func (s *handlerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.msg = &kafka.Message{
		Key:   []byte("enwiki"),
		Value: []byte("Earth"),
	}
	etp := "update"
	art := &schema.Article{
		Name: string(s.msg.Value),
		Event: &schema.Event{
			Type: etp,
		},
	}

	s.stm = new(unmarshalerMock)
	s.stm.On("Unmarshal", s.msg.Value).Return(art, s.sme)

	s.stg = new(storageMock)
	s.stg.On("Update", s.msg.Key, etp, art).Return(s.ste)

	s.prs = &handler.Parameters{
		Stream:  s.stm,
		Storage: s.stg,
		Env:     new(env.Environment),
		Metrics: &prometheus.Metrics{},
	}
}

func (s *handlerTestSuite) TestHandler() {
	hdl := handler.New(s.prs)
	err := hdl(s.ctx, s.msg)

	if s.sme != nil {
		s.Assert().Equal("error unmarshalling message: "+s.sme.Error(), err.Error())
	} else if s.ste != nil {
		s.Assert().Equal("error uploading message: "+s.ste.Error(), err.Error())
	}
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*handlerTestSuite{
		{
			sme: nil,
			ste: nil,
		},
		{
			sme: errors.New("unmarshal failed"),
			ste: nil,
		},
		{
			sme: nil,
			ste: errors.New("update error"),
		},
	} {
		suite.Run(t, testCase)
	}
}
