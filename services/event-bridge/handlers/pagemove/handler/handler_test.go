package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/handlers/pagemove/handler"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"
	"wikimedia-enterprise/services/event-bridge/packages/filter"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	eventstream "github.com/wikimedia-enterprise/wmf-event-stream-sdk-go"
)

var errHandlerTest = errors.New("handler test error")
var errHandlerDictionaryTest = errors.New("handler identifier test error")
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

func (p *handlerProducerMock) Produce(_ context.Context, msgs ...*schema.Message) error {
	for _, msg := range msgs {
		switch m := msg.Value.(type) {
		case *schema.Article:
			if m.Event == nil {
				return fmt.Errorf("empty event field")
			}

			m.Event = nil
			m.DateCreated = nil
			m.DateModified = nil
		}
	}

	return p.Called(msgs[0]).Error(0)
}

type handlerIdentifiersMock struct {
	mock.Mock
	langid.Dictionarer
}

func (d *handlerIdentifiersMock) GetLanguage(_ context.Context, dbname string) (string, error) {
	args := d.Called(dbname)

	return args.Get(0).(string), args.Error(1)
}

type handlerTestSuite struct {
	suite.Suite
	evt        *eventstream.PageMove
	prod       *handlerProducerMock
	redis      *handlerRedisMock
	dictionary *handlerIdentifiersMock
	params     *handler.Parameters
	env        *env.Environment
	filter     *filter.Filter
	expire     time.Duration
	amm        *schema.Message
	adm        *schema.Message
	ctx        context.Context
	key        string
}

func (s *handlerTestSuite) buildArticle(evt *eventstream.PageMove) *schema.Article {
	url, _ := url.QueryUnescape(evt.Data.Meta.URI)
	article := &schema.Article{
		Identifier: evt.Data.PageID,
		Name:       evt.Data.PriorState.PageTitle,
		Version: &schema.Version{
			Identifier: evt.Data.RevID,
			Comment:    evt.Data.Comment,
			Editor: &schema.Editor{
				Identifier:  evt.Data.Performer.UserID,
				Name:        evt.Data.Performer.UserText,
				EditCount:   evt.Data.Performer.UserEditCount,
				Groups:      evt.Data.Performer.UserGroups,
				IsBot:       evt.Data.Performer.UserIsBot,
				DateStarted: &evt.Data.Performer.UserRegistrationDt,
			},
		},
		IsPartOf: &schema.Project{
			Identifier: evt.Data.Database,
			URL:        fmt.Sprintf("https://%s", evt.Data.Meta.Domain),
		},
		Namespace: &schema.Namespace{
			Identifier: evt.Data.PriorState.PageNamespace,
		},
		InLanguage: &schema.Language{
			Identifier: testLangid,
		},
		URL: url,
	}

	return article
}

func (s *handlerTestSuite) SetupSuite() {
	var err error
	s.filter, err = filter.New()
	s.Assert().NoError(err)

	s.env = new(env.Environment)
	s.env.TopicArticleDelete = "local.event-bridge.article-delete.v1"
	s.env.TopicArticleMove = "local.event-bridge.article-move.v1"

	s.key = handler.LastEventTimeKey
	s.expire = time.Hour * 24
}

func (s *handlerTestSuite) SetupTest() {
	article := s.buildArticle(s.evt)

	s.amm = &schema.Message{
		Config: schema.ConfigArticle,
		Topic:  s.env.TopicArticleMove,
		Value:  article,
		Key: &schema.Key{
			Identifier: fmt.Sprintf("/%s/%s", article.IsPartOf.Identifier, article.Name),
			Type:       schema.KeyTypeArticle,
		},
	}

	s.adm = &schema.Message{
		Config: schema.ConfigArticle,
		Topic:  s.env.TopicArticleDelete,
		Value:  article,
		Key: &schema.Key{
			Identifier: fmt.Sprintf("/%s/%s", article.IsPartOf.Identifier, article.Name),
			Type:       schema.KeyTypeArticle,
		},
	}

	s.ctx = context.Background()
	s.prod = new(handlerProducerMock)
	s.redis = new(handlerRedisMock)
	s.dictionary = new(handlerIdentifiersMock)
	s.params = &handler.Parameters{
		Producer:   s.prod,
		Redis:      s.redis,
		Dictionary: s.dictionary,
		Env:        s.env,
	}
}

func (s *handlerTestSuite) TestPageMove() {
	s.prod.On("Produce", s.amm).Return(nil)
	s.prod.On("Produce", s.adm).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(nil)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().NoError(handler.PageMove(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 2)
	s.redis.AssertNumberOfCalls(s.T(), "Set", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestPageMoveRedirect() {
	evt := *s.evt
	evt.Data.PageIsRedirect = true
	s.Assert().NoError(handler.PageMove(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestPageMoveNamespaceFilter() {
	evt := *s.evt
	evt.Data.PageNamespace = 2
	evt.Data.PriorState.PageNamespace = 118

	article := s.buildArticle(&evt)
	amm := s.amm
	amm.Value = article

	s.prod.On("Produce", amm).Return(nil)
	s.redis.On("Set", s.key, evt.Data.Meta.Dt, s.expire).Return(nil)
	s.dictionary.On("GetLanguage", evt.Data.Database).Return(testLangid, nil)
	s.Assert().NoError(handler.PageMove(s.ctx, s.params, s.filter)(&evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
	s.redis.AssertNumberOfCalls(s.T(), "Set", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestPageMoveProjectFilter() {
	evt := *s.evt
	evt.Data.Database = "wikidatawiki"
	s.Assert().NoError(handler.PageMove(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestPageMoveError() {
	s.prod.On("Produce", s.amm).Return(nil)
	s.prod.On("Produce", s.adm).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(errHandlerTest)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().Equal(errHandlerTest, handler.PageMove(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 2)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestPageMoveProducerError() {
	s.prod.On("Produce", s.amm).Return(errHandlerTest)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().Equal(errHandlerTest, handler.PageMove(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestPageMoveRedisError() {
	s.prod.On("Produce", s.amm).Return(nil)
	s.prod.On("Produce", s.adm).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(errHandlerTest)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().Equal(errHandlerTest, handler.PageMove(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 2)
	s.redis.AssertNumberOfCalls(s.T(), "Set", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestDictionaryError() {
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return("", errHandlerDictionaryTest)
	s.Assert().Equal(errHandlerDictionaryTest, handler.PageMove(s.ctx, s.params, s.filter)(s.evt))
}

func TestHandler(t *testing.T) {
	evt := new(eventstream.PageMove)
	evt.Data.PageID = 100
	evt.Data.PageTitle = "Earth"
	evt.Data.Meta.URI = "https://en.wikipedia.com/wiki/Earth"
	evt.Data.Database = "enwiki"
	evt.Data.Performer.UserID = 500
	evt.Data.Performer.UserText = "Andy Murray"
	evt.Data.RevID = 400
	evt.Data.Comment = "Moved page..."
	evt.Data.PageNamespace = 0
	evt.Data.Meta.Dt = time.Now()

	evt.Data.PriorState.PageTitle = "eorthe" // Fun fact: middle english for "earth"
	evt.Data.PriorState.PageNamespace = 0

	for _, testCase := range []*handlerTestSuite{
		{
			evt: evt,
		},
	} {
		suite.Run(t, testCase)
	}
}
