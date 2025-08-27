package operations_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/packages/operations"
	"wikimedia-enterprise/services/eventstream-listener/packages/transformer"
	"wikimedia-enterprise/services/eventstream-listener/submodules/config"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

type operationsNewTestSuite struct {
	suite.Suite
	env *env.Environment
	fr  *filter.Filter
	trs *transformer.Transforms
}

func (s *operationsNewTestSuite) SetupSuite() {
	s.env = new(env.Environment)

	cfg, err := config.New()
	s.Assert().NoError(err)

	s.fr = filter.New(cfg)
	s.trs = transformer.New(s.fr)
}

func (s *operationsNewTestSuite) TestNew() {
	s.Assert().NotNil(operations.New(s.trs, s.fr, s.env))
}

func TestNew(t *testing.T) {
	suite.Run(t, new(operationsNewTestSuite))
}

type operationMock struct {
	mock.Mock
	operations.Operation
}

func (o *operationMock) Execute(_ context.Context, _ *eventstream.PageChange) ([]*schema.Message, error) {
	ags := o.Called()

	return ags.Get(0).([]*schema.Message), ags.Error(1)
}

type operationsExecuteTestSuite struct {
	suite.Suite
	env   *env.Environment
	fr    *filter.Filter
	trs   *transformer.Transforms
	ops   *operations.Operations
	msgs  []*schema.Message
	event *eventstream.PageChange
	err   error
	ctx   context.Context
}

func (s *operationsExecuteTestSuite) SetupSuite() {
	s.env = new(env.Environment)

	cfg, err := config.New()
	s.Assert().NoError(err)

	s.fr = filter.New(cfg)
	s.trs = transformer.New(s.fr)

	s.ops = operations.New(s.trs, s.fr, s.env)
	s.ctx = context.Background()
}

func (s *operationsExecuteTestSuite) TestExecute() {
	operation := new(operationMock)
	s.ops.OperationsMap["create"] = operation

	operation.On("Execute").Return(s.msgs, s.err)

	msgs, err := s.ops.Execute(s.ctx, s.event)
	s.Assert().ElementsMatch(msgs, s.msgs)

	if s.err == nil {
		s.Assert().NoError(err)
	} else {
		s.Assert().Equal(err, s.err)
	}
}

func TestExecute(t *testing.T) {
	testEventA := new(eventstream.PageChange)
	testEventA.Data.PageChangeKind = "create"

	testEventB := new(eventstream.PageChange)
	testEventB.Data.PageChangeKind = "unsupported"

	for _, testcase := range []*operationsExecuteTestSuite{
		{
			msgs:  []*schema.Message{{Topic: "aws.event-bridge.article-create.v1"}},
			err:   nil,
			event: testEventA,
		},
		{
			msgs:  []*schema.Message{},
			err:   errors.New("unsupported event type unsupported"),
			event: testEventB,
		},
		{
			msgs:  []*schema.Message{{Topic: "aws.event-bridge.article-create.v1"}},
			err:   errors.New("operation error"),
			event: testEventA,
		},
	} {
		suite.Run(t, testcase)
	}
}
