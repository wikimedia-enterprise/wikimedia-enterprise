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

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

type moveNewTestSuite struct {
	suite.Suite
	env *env.Environment
	fr  *filter.Filter
	trs *transformer.Transforms
}

func (s *moveNewTestSuite) SetupSuite() {
	s.env = new(env.Environment)

	cfg, err := config.New()
	s.Assert().NoError(err)

	s.fr = filter.New(cfg)
	s.trs = transformer.New(s.fr)
}

func (s *moveNewTestSuite) TestNew() {
	s.Assert().NotNil(operations.NewMoveOperation(s.trs, s.fr, s.env))
}

func TestMoveNew(t *testing.T) {
	suite.Run(t, new(moveNewTestSuite))
}

type moveExecuteTestSuite struct {
	suite.Suite
	env      *env.Environment
	fr       *filter.Filter
	trs      *transformer.Transforms
	move     operations.Operation
	msgCount int
	topics   []string
	event    *eventstream.PageChange
	ctx      context.Context
	err      error
}

func (s *moveExecuteTestSuite) SetupTest() {
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
	s.move = operations.NewMoveOperation(s.trs, s.fr, s.env)
}

func (s *moveExecuteTestSuite) TestExecute() {
	msgs, err := s.move.Execute(s.ctx, s.event)
	s.Assert().Len(msgs, s.msgCount)

	if s.err == nil {
		s.Assert().NoError(err)
	} else {
		s.Assert().Equal(err.Error(), s.err.Error())
	}

	topics := []string{}

	if s.msgCount > 0 {
		for _, msg := range msgs {
			topics = append(topics, msg.Topic)
		}
	}

	s.Assert().ElementsMatch(s.topics, topics)
}

func TestMoveExecute(t *testing.T) {
	testEventA := new(eventstream.PageChange)
	testEventA.Data.Database = "unsupported"

	testEventB := new(eventstream.PageChange)
	testEventB.Data.Database = "enwiki"
	testEventB.Data.Page.PageNamespace = -9999999
	testEventB.Data.PriorState.Page.PageNamespace = -9999999

	testEventC := new(eventstream.PageChange)
	testEventC.Data.Database = "enwiki"
	testEventC.Data.Page.PageNamespace = 0
	testEventC.Data.PriorState.Page.PageNamespace = -9999999
	testEventC.Data.Page.PageIsRedirect = true

	testEventD := new(eventstream.PageChange)
	testEventD.Data.Database = "enwiki"
	testEventD.Data.Page.PageNamespace = 0
	testEventD.Data.PriorState.Page.PageNamespace = 6
	testEventD.Data.Page.PageIsRedirect = true

	testEventE := new(eventstream.PageChange)
	testEventE.Data.Database = "commonswiki"
	testEventE.Data.Page.PageNamespace = 6
	testEventE.Data.PriorState.Page.PageNamespace = 6
	testEventE.Data.Page.PageIsRedirect = false

	testEventF := new(eventstream.PageChange)
	testEventF.Data.Database = "enwiki"
	testEventF.Data.Page.PageNamespace = 6
	testEventF.Data.PriorState.Page.PageNamespace = 14
	testEventF.Data.Page.PageIsRedirect = false

	for _, testcase := range []*moveExecuteTestSuite{
		{
			msgCount: 0,
			event:    testEventA,
			err:      errors.New("could not find language for project unsupported"),
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
			topics:   []string{"aws.event-bridge.article-delete.v1"},
		},
		{
			msgCount: 2,
			event:    testEventE,
			topics:   []string{"aws.event-bridge.commons-delete.v1", "aws.event-bridge.commons-move.v1"},
		},
		{
			msgCount: 2,
			event:    testEventF,
			topics:   []string{"aws.event-bridge.article-delete.v1", "aws.event-bridge.article-move.v1"},
		},
	} {
		suite.Run(t, testcase)
	}
}
