package aggregate

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/services/content-integrity/config/env"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"
	"wikimedia-enterprise/services/content-integrity/submodules/schema"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type collectorMock struct {
	mock.Mock
	collector.ArticleAPI
}

func (mc *collectorMock) SetDateCreated(_ context.Context, _ string, _ int, _ *time.Time) (*time.Time, error) {
	return nil, mc.Called().Error(1)
}

func (mc *collectorMock) SetDateNamespaceMoved(_ context.Context, _ string, _ int, _ *time.Time) (*time.Time, error) {
	return nil, mc.Called().Error(1)
}

func (mc *collectorMock) PrependVersion(_ context.Context, _ string, _ int, _ *collector.Version) (collector.Versions, error) {
	return nil, mc.Called().Error(1)
}

func (mc *collectorMock) SetName(_ context.Context, _ string, _ int, _ string) error {
	return mc.Called().Error(1)
}

type streamMock struct {
	schema.UnmarshalProducer
	mock.Mock
}

func (m *streamMock) Unmarshal(_ context.Context, data []byte, v interface{}) error {
	ags := m.Called(data)

	if v != nil {
		*v.(*schema.Article) = *ags.Get(0).(*schema.Article)
	}

	return ags.Error(1)
}

type aggregateTestSuite struct {
	suite.Suite
	ctx context.Context
	env *env.Environment
	prs *Parameters
	msg *kafka.Message
	clr *collectorMock
	stm *streamMock
	art *schema.Article
}

func (s *aggregateTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.env = &env.Environment{}
	s.msg = &kafka.Message{
		Key:   []byte("enwiki"),
		Value: []byte("Earth"),
	}
}

func (s *aggregateTestSuite) SetupTest() {
	tmn := time.Now()

	s.art = &schema.Article{
		Name: string(s.msg.Key),
		Event: &schema.Event{
			Type: schema.EventTypeCreate,
		},
		Namespace: &schema.Namespace{
			Name: "enwiki",
		},
		IsPartOf: &schema.Project{
			Identifier: "enwiki",
		},
		Identifier:  100,
		DateCreated: &tmn,
		Version: &schema.Version{
			Identifier: 200,
			Editor: &schema.Editor{
				Name: "test",
			},
		},
	}
	s.stm = new(streamMock)
	s.clr = new(collectorMock)
	s.prs = &Parameters{
		Stream:    s.stm,
		Collector: s.clr,
	}
}

func (s *aggregateTestSuite) TestHandler() {
	s.stm.On("Unmarshal", s.msg.Value).Return(s.art, nil)
	s.clr.On("SetDateCreated").Return(nil, nil)
	s.clr.On("PrependVersion").Return(nil, nil)
	s.clr.On("SetDateNamespaceMoved").Return(nil, nil)
	s.clr.On("SetName").Return("", nil)

	hdl := NewAggregate(s.prs)
	err := hdl(s.ctx, s.msg)
	s.Assert().NoError(err)

	s.art.Event.Type = schema.EventTypeCreate
	err = hdl(s.ctx, s.msg)
	s.Assert().NoError(err)

	s.art.Event.Type = schema.EventTypeUpdate
	err = hdl(s.ctx, s.msg)
	s.Assert().NoError(err)

	s.art.Event.Type = schema.EventTypeMove
	err = hdl(s.ctx, s.msg)
	s.Assert().NoError(err)
}

func (s *aggregateTestSuite) TestUnmarshalAggregateError() {
	uer := errors.New("unmarshal error")
	s.stm.On("Unmarshal", s.msg.Value).Return(s.art, uer)

	hdl := NewAggregate(s.prs)
	err := hdl(s.ctx, s.msg)
	s.Assert().Error(err)
	s.Assert().Equal(uer, err)
}

func (s *aggregateTestSuite) TestSetDateCreatedErrors() {
	s.stm.On("Unmarshal", s.msg.Value).Return(s.art, nil)

	dce := errors.New("SetDateCreated error")
	s.clr.On("SetDateCreated").Return(nil, dce)

	hdl := NewAggregate(s.prs)
	err := hdl(s.ctx, s.msg)

	s.Assert().Error(err)
	s.Assert().Equal(dce, err)
}

func (s *aggregateTestSuite) TestPrependVersionErrors() {
	s.stm.On("Unmarshal", s.msg.Value).Return(s.art, nil)

	pve := errors.New("PrependVersion error")
	s.art.Event.Type = schema.EventTypeUpdate
	s.clr.On("PrependVersion").Return(nil, pve)

	hdl := NewAggregate(s.prs)
	err := hdl(s.ctx, s.msg)

	s.Assert().Error(err)
	s.Assert().Equal(pve, err)
}

func (s *aggregateTestSuite) TestSetDateNamespaceMovedErrors() {
	s.stm.On("Unmarshal", s.msg.Value).Return(s.art, nil)

	nme := errors.New("SetDateNamespaceMoved error")
	s.art.Event.Type = schema.EventTypeMove
	s.clr.On("SetDateNamespaceMoved").Return(nil, nme)

	hdl := NewAggregate(s.prs)
	err := hdl(s.ctx, s.msg)

	s.Assert().Error(err)
	s.Assert().Equal(nme, err)
}

func TestAggregate(t *testing.T) {
	suite.Run(t, new(aggregateTestSuite))
}
