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

type visibilityNewTestSuite struct {
	suite.Suite
	env *env.Environment
	fr  *filter.Filter
	trs *transformer.Transforms
}

func (s *visibilityNewTestSuite) SetupSuite() {
	s.env = new(env.Environment)

	cfg, err := config.New()
	s.Assert().NoError(err)

	s.fr = filter.New(cfg)
	s.trs = transformer.New(s.fr)
}

func (s *visibilityNewTestSuite) TestNew() {
	s.Assert().NotNil(operations.NewVisibilityOperation(s.trs, s.fr, s.env))
}

func TestVisibilityNew(t *testing.T) {
	suite.Run(t, new(visibilityNewTestSuite))
}

type visibilityExecuteTestSuite struct {
	suite.Suite
	env      *env.Environment
	fr       *filter.Filter
	trs      *transformer.Transforms
	vis      operations.Operation
	msgCount int
	topic    string
	event    *eventstream.PageChange
	ctx      context.Context
}

func (s *visibilityExecuteTestSuite) SetupTest() {
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
	s.vis = operations.NewVisibilityOperation(s.trs, s.fr, s.env)
}

func (s *visibilityExecuteTestSuite) TestExecute() {
	msgs, err := s.vis.Execute(s.ctx, s.event)
	s.Assert().Len(msgs, s.msgCount)
	s.Assert().NoError(err)

	if s.msgCount > 0 {
		for _, msg := range msgs {
			s.Assert().Equal(msg.Topic, s.topic)
			art := msg.Value.(*schema.Article)
			art.Visibility.Comment = s.event.Data.Revision.IsCommentVisible
			art.Visibility.Editor = s.event.Data.Revision.IsEditorVisible
			art.Visibility.Text = s.event.Data.Revision.IsContentVisible
		}
	}
}

func TestVisExecute(t *testing.T) {
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
	testEventD.Data.Revision.IsCommentVisible = true
	testEventD.Data.Revision.IsContentVisible = true
	testEventD.Data.Revision.IsEditorVisible = false

	testEventE := new(eventstream.PageChange)
	testEventE.Data.Database = "commonswiki"
	testEventE.Data.Page.PageNamespace = 6
	testEventE.Data.Page.PageIsRedirect = false
	testEventE.Data.Revision.IsCommentVisible = false
	testEventE.Data.Revision.IsContentVisible = true
	testEventE.Data.Revision.IsEditorVisible = false

	for _, testcase := range []*visibilityExecuteTestSuite{
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
			topic:    "aws.event-bridge.article-visibility-change.v1",
		},
		{
			msgCount: 1,
			event:    testEventE,
			topic:    "aws.event-bridge.commons-visibility-change.v1",
		},
	} {
		suite.Run(t, testcase)
	}
}
