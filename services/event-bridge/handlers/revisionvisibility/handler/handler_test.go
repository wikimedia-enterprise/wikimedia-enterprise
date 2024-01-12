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
	"wikimedia-enterprise/services/event-bridge/handlers/revisionvisibility/handler"
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
	evt        *eventstream.RevisionVisibilityChange
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
	s.env.TopicArticleVisibilityChange = "local.event-bridge.article-visibility.v1"

	s.key = handler.LastEventTimeKey
	s.expire = time.Hour * 24
}

func (s *handlerTestSuite) SetupTest() {
	url, _ := url.QueryUnescape(s.evt.Data.Meta.URI)
	article := new(schema.Article)
	article.Identifier = s.evt.Data.PageID
	article.Name = s.evt.Data.PageTitle
	article.DateModified = &s.evt.Data.RevTimestamp
	article.PreviousVersion = &schema.PreviousVersion{
		Identifier: s.evt.Data.RevParentID,
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
	article.IsPartOf = &schema.Project{
		Identifier: s.evt.Data.Database,
		URL:        fmt.Sprintf("https://%s", s.evt.Data.Meta.Domain),
	}
	article.Visibility = &schema.Visibility{
		Text:    s.evt.Data.Visibility.Text,
		Editor:  s.evt.Data.Visibility.User,
		Comment: s.evt.Data.Visibility.Comment,
	}
	article.Namespace = &schema.Namespace{
		Identifier: s.evt.Data.PageNamespace,
	}
	article.InLanguage = &schema.Language{
		Identifier: testLangid,
	}
	article.URL = url

	s.msg = &schema.Message{
		Config: schema.ConfigArticle,
		Topic:  s.env.TopicArticleVisibilityChange,
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

func (s *handlerTestSuite) TestRevisionVisibility() {
	s.prod.On("Produce", s.msg).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(nil)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().NoError(handler.RevisionVisibility(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
	s.redis.AssertNumberOfCalls(s.T(), "Set", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestRevisionVisibilityRedirect() {
	evt := *s.evt
	evt.Data.PageIsRedirect = true
	s.Assert().NoError(handler.RevisionVisibility(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestRevisionVisibilityNamespaceFilter() {
	evt := *s.evt
	evt.Data.PageNamespace = 2
	s.prod.On("Produce", s.msg).Return(nil)
	s.Assert().NoError(handler.RevisionVisibility(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestRevisionVisibilityProjectFilter() {
	evt := *s.evt
	evt.Data.Database = "wikidatawiki"
	s.prod.On("Produce", s.msg).Return(nil)
	s.Assert().NoError(handler.RevisionVisibility(s.ctx, s.params, s.filter)(&evt))
}

func (s *handlerTestSuite) TestRevisionVisibilityProducerError() {
	s.prod.On("Produce", s.msg).Return(errHandlerTest)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().Equal(errHandlerTest, handler.RevisionVisibility(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
}

func (s *handlerTestSuite) TestRevisionVisibilityRedisError() {
	s.prod.On("Produce", s.msg).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(errHandlerTest)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return(testLangid, nil)
	s.Assert().Equal(errHandlerTest, handler.RevisionVisibility(s.ctx, s.params, s.filter)(s.evt))
	s.prod.AssertNumberOfCalls(s.T(), "Produce", 1)
	s.redis.AssertNumberOfCalls(s.T(), "Set", 1)
	s.dictionary.AssertNumberOfCalls(s.T(), "GetLanguage", 1)
}

func (s *handlerTestSuite) TestDictionaryError() {
	s.prod.On("Produce", s.msg).Return(nil)
	s.redis.On("Set", s.key, s.evt.Data.Meta.Dt, s.expire).Return(nil)
	s.dictionary.On("GetLanguage", s.evt.Data.Database).Return("", errHandlerDictionaryTest)
	s.Assert().Equal(errHandlerDictionaryTest, handler.RevisionVisibility(s.ctx, s.params, s.filter)(s.evt))
}

func TestHandler(t *testing.T) {
	dateTest := time.Date(2015, 8, 11, 20, 34, 58, 651387237, time.UTC)
	evt := new(eventstream.RevisionVisibilityChange)
	evt.Data.PageID = 1001
	evt.Data.PageTitle = "Monty_Python"
	evt.Data.RevTimestamp = dateTest
	evt.Data.Meta.URI = "https://en.wikipedia.com/wiki/Earth"
	evt.Data.RevID = 2002
	evt.Data.RevParentID = 3003
	evt.Data.Comment = "test comment"
	evt.Data.Performer.UserID = 555
	evt.Data.Performer.UserText = "Arthur"
	evt.Data.Performer.UserEditCount = 16582
	evt.Data.Performer.UserGroups = []string{"sysop", "user"}
	evt.Data.Performer.UserIsBot = false
	evt.Data.Performer.UserRegistrationDt = dateTest
	evt.Data.Visibility.Text = false
	evt.Data.Visibility.User = true
	evt.Data.Visibility.Comment = true
	evt.Data.PageNamespace = 0
	evt.Data.Database = "enwiki"

	for _, testCase := range []*handlerTestSuite{
		{
			evt: evt,
		},
	} {
		suite.Run(t, testCase)
	}
}
