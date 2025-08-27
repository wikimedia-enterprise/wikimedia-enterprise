package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/handler"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/packages/operations"
	"wikimedia-enterprise/services/eventstream-listener/packages/transformer"
	"wikimedia-enterprise/services/eventstream-listener/submodules/config"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

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

func (p *handlerProducerMock) Produce(ctx context.Context, msgs ...*schema.Message) error {
	return p.Called(ctx, msgs).Error(0)
}

type handlerTestSuite struct {
	suite.Suite
	evt      *eventstream.PageChange
	prod     *handlerProducerMock
	redis    *handlerRedisMock
	fr       *filter.Filter
	trs      *transformer.Transforms
	ops      *operations.Operations
	params   *handler.Parameters
	env      *env.Environment
	expire   time.Duration
	ctx      context.Context
	key      string
	redisErr error
	prodErr  error
	operErr  error
	msgs     []*schema.Message
	hdl      func(evt *eventstream.PageChange) error
}

type tracerMock struct{}

func (t *tracerMock) Trace(ctx context.Context, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

func (t *tracerMock) Shutdown(ctx context.Context) error {
	return nil
}

func (t *tracerMock) StartTrace(ctx context.Context, _ string, _ map[string]string) (func(err error, msg string), context.Context) {
	return func(err error, msg string) {}, ctx
}

type operationMock struct {
	mock.Mock
	operations.Operation
}

func (o *operationMock) Execute(_ context.Context, _ *eventstream.PageChange) ([]*schema.Message, error) {
	ags := o.Called()

	return ags.Get(0).([]*schema.Message), ags.Error(1)
}

func (s *handlerTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.OutputTopics = &schema.Topics{
		Versions:    []string{"v1"},
		ServiceName: "event-bridge",
		Location:    "aws",
	}
	cfg, err := config.New()
	s.Assert().NoError(err)

	s.fr = filter.New(cfg)
	s.key = handler.LastEventTimeKey
	s.expire = time.Hour * 24
	s.ctx = context.Background()
	s.trs = transformer.New(s.fr)
	s.ops = operations.New(s.trs, s.fr, s.env)
	s.msgs = []*schema.Message{}
}

func (s *handlerTestSuite) SetupTest() {

}

func (s *handlerTestSuite) TestPageChange() {
	operation := new(operationMock)
	s.ops.OperationsMap["create"] = operation
	s.prod = new(handlerProducerMock)
	s.redis = new(handlerRedisMock)

	s.params = &handler.Parameters{
		Producer:   s.prod,
		Redis:      s.redis,
		Env:        s.env,
		Tracer:     new(tracerMock),
		Filter:     s.fr,
		Operations: s.ops,
	}
	operation.On("Execute").Return(s.msgs, s.operErr)
	s.prod.On("Produce", mock.Anything, mock.Anything).Return(s.prodErr)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(s.redisErr)

	s.hdl = handler.PageChange(s.ctx, s.params)
	err := s.hdl(s.evt)

	if s.prodErr != nil {
		s.Assert().Equal(err, s.prodErr)
		return
	} else if s.redisErr != nil {
		s.Assert().Equal(err, s.redisErr)
		return
	} else if s.operErr != nil {
		s.Assert().Equal(err, s.operErr)
		return
	}

	s.Assert().NoError(err)
}

func TestHandler(t *testing.T) {
	evtA := new(eventstream.PageChange)
	evtA.Data.Page.PageID = 100
	evtA.Data.Meta.URI = "https://en.wikipedia.com/wiki/Earth"
	evtA.Data.Page.PageIsRedirect = false
	evtA.Data.Page.PageTitle = "Earth"
	evtA.Data.Revision.RevID = 200
	evtA.Data.Performer.UserID = 400
	evtA.Data.PageChangeKind = operations.Create
	evtA.Data.Page.PageNamespace = 6
	evtA.Data.Database = "commonswiki"

	evtB := *evtA
	evtB.Data.Page.PageNamespace = -6
	evtB.Data.PriorState.Page.PageNamespace = -6

	evtC := *evtA
	evtB.Data.Database = "unsupported"

	for _, testCase := range []*handlerTestSuite{
		{
			evt: evtA,
		},
		{
			evt: &evtB,
		},
		{
			evt: &evtC,
		},
		{
			evt:     evtA,
			prodErr: errors.New("producer error"),
		},
		{
			evt:      evtA,
			redisErr: errors.New("redis set error"),
		},
		{
			evt:     evtA,
			operErr: errors.New("operation error"),
		},
	} {
		suite.Run(t, testCase)
	}
}
