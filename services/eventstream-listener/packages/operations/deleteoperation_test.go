package operations_test

import (
	"context"
	"testing"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/packages/filter"
	"wikimedia-enterprise/services/eventstream-listener/packages/operations"
	"wikimedia-enterprise/services/eventstream-listener/packages/transformer"
	"wikimedia-enterprise/services/eventstream-listener/submodules/config"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

type deleteNewTestSuite struct {
	suite.Suite
	env *env.Environment
	fr  *filter.Filter
	trs *transformer.Transforms
}

func (s *deleteNewTestSuite) SetupSuite() {
	s.env = new(env.Environment)

	cfg, err := config.New()
	s.Assert().NoError(err)

	s.fr = filter.New(cfg)
	s.trs = transformer.New(s.fr)
}

func (s *deleteNewTestSuite) TestNew() {
	s.Assert().NotNil(operations.NewDeleteOperation(s.trs, s.fr, s.env))
}

func TestDeleteNew(t *testing.T) {
	suite.Run(t, new(deleteNewTestSuite))
}

type deleteExecuteTestSuite struct {
	suite.Suite
	env      *env.Environment
	fr       *filter.Filter
	trs      *transformer.Transforms
	delete   operations.Operation
	msgCount int
	topic    string
	event    *eventstream.PageChange
	ctx      context.Context
}

func (s *deleteExecuteTestSuite) SetupTest() {
	s.env = new(env.Environment)
	s.env.OutputTopics = &schema.Topics{
		Versions:    []string{"v1"},
		ServiceName: "event-bridge",
		Location:    "aws",
	}

	cfg, err := config.New()
	s.Assert().NoError(err)

	s.fr = filter.New(cfg)
	s.trs = transformer.New(s.fr)

	s.ctx = context.WithValue(context.Background(), operations.EventIdentifierKey, uuid.New().String())
	s.delete = operations.NewDeleteOperation(s.trs, s.fr, s.env)
}

func (s *deleteExecuteTestSuite) TestExecute() {
	msgs, err := s.delete.Execute(s.ctx, s.event)
	s.Assert().Len(msgs, s.msgCount)
	s.Assert().NoError(err)

	if s.msgCount > 0 {
		for _, msg := range msgs {
			s.Assert().Equal(msg.Topic, s.topic)
		}
	}
}

func TestDeleteExecute(t *testing.T) {
	testEventA := new(eventstream.PageChange)
	testEventA.Data.Database = "unsupported"

	testEventB := new(eventstream.PageChange)
	testEventB.Data.Database = "enwiki"
	testEventB.Data.Page.PageNamespace = -9999999

	testEventC := new(eventstream.PageChange)
	testEventC.Data.Database = "enwiki"
	testEventC.Data.Page.PageNamespace = 0
	testEventC.Data.Page.PageIsRedirect = true

	testEventD := new(eventstream.PageChange)
	testEventD.Data.Database = "enwiki"
	testEventD.Data.Page.PageNamespace = 0
	testEventD.Data.Page.PageIsRedirect = false

	testEventE := new(eventstream.PageChange)
	testEventE.Data.Database = "commonswiki"
	testEventE.Data.Page.PageNamespace = 6
	testEventE.Data.Page.PageIsRedirect = false

	for _, testcase := range []*deleteExecuteTestSuite{
		{
			msgCount: 0,
			event:    testEventA,
		},
		{
			msgCount: 0,
			event:    testEventB,
		},
		{
			msgCount: 0,
			event:    testEventC,
		},
		{
			msgCount: 1,
			event:    testEventD,
			topic:    "aws.event-bridge.article-delete.v1",
		},
		{
			msgCount: 1,
			event:    testEventE,
			topic:    "aws.event-bridge.commons-delete.v1",
		},
	} {
		suite.Run(t, testcase)
	}
}
