package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/handlers/pagecreate/handler"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/filter"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

var testLangid = "en"

type handlerRedisMock struct {
	mock.Mock
	redis.Cmdable
}

func (r *handlerRedisMock) Set(_ context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := r.Called(key, value, expiration)
	cmd := new(redis.StatusCmd)
	cmd.SetErr(args.Error(0))
	return cmd
}

type handlerProducerMock struct {
	mock.Mock
	schema.UnmarshalProducer
}

func (p *handlerProducerMock) Produce(_ context.Context, _ ...*schema.Message) error {
	return p.Called().Error(0)
}

type handlerIdentifiersMock struct {
	langid.Dictionarer
	mock.Mock
}

func (d *handlerIdentifiersMock) GetLanguage(_ context.Context, dbname string) (string, error) {
	args := d.Called(dbname)

	return args.Get(0).(string), args.Error(1)
}

type handlerTestSuite struct {
	suite.Suite
	evt        *eventstream.PageCreate
	prod       *handlerProducerMock
	redis      *handlerRedisMock
	dictionary *handlerIdentifiersMock
	params     *handler.Parameters
	env        *env.Environment
	filter     *filter.Filter
	expire     time.Duration
	ctx        context.Context
	key        string
	langErr    error
	redisErr   error
	prodErr    error
	hdl        func(evt *eventstream.PageCreate) error
}

func (s *handlerTestSuite) SetupSuite() {
	var err error
	s.filter, err = filter.New()
	s.Assert().NoError(err)

	s.env = new(env.Environment)
	s.env.TopicArticleCreate = "local.event-bridge.article-create.v1"

	s.key = handler.LastEventTimeKey
	s.expire = time.Hour * 24
	s.ctx = context.Background()
}

func (s *handlerTestSuite) SetupTest() {
	s.prod = new(handlerProducerMock)
	s.redis = new(handlerRedisMock)
	s.dictionary = new(handlerIdentifiersMock)
	s.params = &handler.Parameters{
		Producer:   s.prod,
		Redis:      s.redis,
		Env:        s.env,
		Dictionary: s.dictionary,
	}

	s.hdl = handler.PageCreate(s.ctx, s.params, s.filter)
}

func (s *handlerTestSuite) TestPageCreate() {
	s.prod.On("Produce").Return(s.prodErr)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(s.redisErr)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, s.langErr)

	err := s.hdl(s.evt)

	if s.prodErr != nil {
		s.Assert().Equal(err, s.prodErr)
	} else if s.redisErr != nil {
		s.Assert().Equal(err, s.redisErr)
	} else if s.langErr != nil {
		s.Assert().Equal(err, s.langErr)
	} else {
		s.Assert().NoError(err)
	}
}

func TestHandler(t *testing.T) {
	evtA := new(eventstream.PageCreate)
	evtA.Data.PageID = 100
	evtA.Data.Meta.URI = "https://en.wikipedia.com/wiki/Earth"
	evtA.Data.PageIsRedirect = false
	evtA.Data.PageTitle = "Earth"
	evtA.Data.RevID = 200
	evtA.Data.Performer.UserID = 400
	evtA.Data.PageNamespace = 0
	evtA.Data.Database = "enwiki"

	evtB := *evtA
	evtB.Data.PageIsRedirect = true

	evtC := *evtA
	evtB.Data.Database = "unsupported"

	for _, testCase := range []*handlerTestSuite{
		{
			evt:      evtA,
			langErr:  nil,
			prodErr:  nil,
			redisErr: nil,
		},
		{
			evt:      &evtB,
			langErr:  nil,
			prodErr:  nil,
			redisErr: nil,
		},
		{
			evt:      &evtC,
			langErr:  nil,
			prodErr:  nil,
			redisErr: nil,
		},
		{
			evt:      evtA,
			langErr:  errors.New("lang error"),
			prodErr:  nil,
			redisErr: nil,
		},
		{
			evt:      evtA,
			langErr:  nil,
			prodErr:  errors.New("producer error"),
			redisErr: nil,
		},
		{
			evt:      evtA,
			langErr:  nil,
			prodErr:  errors.New("redis set error"),
			redisErr: nil,
		},
	} {
		suite.Run(t, testCase)
	}
}
