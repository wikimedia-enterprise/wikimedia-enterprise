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
	"wikimedia-enterprise/services/event-bridge/handlers/pagedelete/handler"
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
				return fmt.Errorf("empty envent field")
			}

			m.Event = nil
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
	evt        *eventstream.PageDelete
	prod       *handlerProducerMock
	redis      *handlerRedisMock
	dictionary *handlerIdentifiersMock
	params     *handler.Parameters
	env        *env.Environment
	filter     *filter.Filter
	expire     time.Duration
	msg        *schema.Message
	ctx        context.Context
	key        string
}

func (s *handlerTestSuite) SetupSuite() {
	var err error
	s.filter, err = filter.New()
	s.Assert().NoError(err)

	s.env = new(env.Environment)
	s.env.TopicArticleDelete = "local.event-bridge.article-delete.v1"

	s.key = handler.LastEventTimeKey
	s.expire = time.Hour * 24
}

func (s *handlerTestSuite) SetupTest() {
	url, _ := url.QueryUnescape(s.evt.Data.Meta.URI)
	article := new(schema.Article)
	article.Name = s.evt.Data.PageTitle
	article.Identifier = s.evt.Data.PageID
	article.Namespace = &schema.Namespace{
		Identifier: s.evt.Data.PageNamespace,
	}
	article.IsPartOf = &schema.Project{
		Identifier: s.evt.Data.Database,
		URL:        fmt.Sprintf("https://%s", s.evt.Data.Meta.Domain),
	}
	article.InLanguage = &schema.Language{
		Identifier: testLangid,
	}
	article.Version = &schema.Version{
		Identifier: s.evt.Data.RevID,
		Comment:    s.evt.Data.Comment,
		Editor: &schema.Editor{
			Identifier:  s.evt.Data.Performer.UserID,
			Name:        s.evt.Data.Performer.UserText,
			EditCount:   s.evt.Data.Performer.UserEditCount,
			Groups:      s.evt.Data.Performer.UserGroups,
			IsBot:       s.evt.Data.Performer.UserIsBot,
			DateStarted: &s.evt.Data.Performer.UserRegistrationDt,
		},
	}

	article.URL = url

	s.msg = &schema.Message{
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
		Env:        s.env,
		Dictionary: s.dictionary,
	}
}

func (s *handlerTestSuite) TestPageDelete() {
	s.prod.On("Produce", s.msg).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(nil)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().NoError(handler.PageDelete(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
	s.redis.AssertNumberOfCalls(s.T(), "Set", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestPageDeleteRedirect() {
	evt := *s.evt
	evt.Data.PageIsRedirect = true
	s.Assert().NoError(handler.PageDelete(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestPageDeleteNamespaceFilter() {
	evt := *s.evt
	evt.Data.PageNamespace = 2
	s.Assert().NoError(handler.PageDelete(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestPageDeleteProjectFilter() {
	evt := *s.evt
	evt.Data.Database = "wikidatawiki"
	s.Assert().NoError(handler.PageDelete(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestPageDeleteProducerError() {
	s.prod.On("Produce", s.msg).Return(errHandlerTest)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().Equal(errHandlerTest, handler.PageDelete(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestPageDeleteRedisError() {
	s.prod.On("Produce", s.msg).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(errHandlerTest)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().Equal(errHandlerTest, handler.PageDelete(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
	s.redis.AssertNumberOfCalls(s.T(), "Set", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestDictionaryError() {
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return("", errHandlerDictionaryTest)
	s.Assert().Equal(errHandlerDictionaryTest, handler.PageDelete(s.ctx, s.params, s.filter)(s.evt))
}

func TestHandler(t *testing.T) {
	evt := new(eventstream.PageDelete)
	evt.Data.Meta.Domain = "en.wikipedia.com"
	evt.Data.Meta.URI = "https://en.wikipedia.com/wiki/Earth"
	evt.Data.PageID = 100
	evt.Data.PageTitle = "Earth"
	evt.Data.Database = "enwiki"
	evt.Data.Performer.UserID = 500
	evt.Data.Performer.UserText = "Andy Murray"
	evt.Data.RevID = 400
	evt.Data.Comment = "The dog chewed this page...."
	evt.Data.PageNamespace = 0
	evt.Data.Meta.Dt = time.Now()

	for _, testCase := range []*handlerTestSuite{
		{
			evt: evt,
		},
	} {
		suite.Run(t, testCase)
	}
}
