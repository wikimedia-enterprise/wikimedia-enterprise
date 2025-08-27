package articles_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"wikimedia-enterprise/api/realtime/config/env"
	articles "wikimedia-enterprise/api/realtime/handlers/v2/articles"
	libkafka "wikimedia-enterprise/api/realtime/libraries/kafka"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/schema"
)

// mockConsumer implements ReadAllCloser for testing.
type mockConsumer struct {
	mock.Mock
	msgData []byte
}

func (m *mockConsumer) ReadAll(ctx context.Context, params *libkafka.ReadParams, topic string, partitions []int, cb func(msg *kafka.Message) error) error {
	ags := m.Called(ctx, params, topic, partitions, cb)
	return ags.Error(0)
}

func (m *mockConsumer) SetPartitionOffsets(topic string, partitions []int, params *libkafka.ReadParams) ([]kafka.TopicPartition, error) {
	tps := make([]kafka.TopicPartition, len(partitions))
	for i, p := range partitions {
		tps[i] = kafka.TopicPartition{Topic: &topic, Partition: int32(p), Offset: kafka.OffsetEnd}
	}
	return tps, nil
}

func (m *mockConsumer) Close() error {
	return m.Called().Error(0)
}

// mockConsumerGetter returns the same mockConsumer instance.
type mockConsumerGetter struct {
	consumer libkafka.ReadAllCloser
	mock.Mock
}

func (m *mockConsumerGetter) GetConsumer(_ string) (libkafka.ReadAllCloser, error) {
	ags := m.Called()
	return ags.Get(0).(libkafka.ReadAllCloser), nil
}

type mockUnmarshaler struct{}

func (u *mockUnmarshaler) Unmarshal(_ context.Context, data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (u *mockUnmarshaler) UnmarshalNoCache(ctx context.Context, data []byte, v interface{}) error {
	return nil
}

type handlerTestSuite struct {
	suite.Suite
	srv        *httptest.Server
	readAllErr error
	getConErr  error
	closeErr   error
	reqParams  *articles.Model
	status     int
	topic      string
	partitions []int
	msg        []byte
	unmar      schema.Unmarshaler
}

func (s *handlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.unmar = new(mockUnmarshaler)
}

func (s *handlerTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *handlerTestSuite) TestNewHandler() {
	testArticle := schema.Article{Name: "TestArticle"}
	var err error
	s.msg, err = json.Marshal(testArticle)
	s.Require().NoError(err)
	mc := &mockConsumer{msgData: s.msg}
	cg := &mockConsumerGetter{consumer: mc}

	mc.On("ReadAll", mock.Anything, mock.Anything, s.topic, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		cb := args.Get(4).(func(*kafka.Message) error)

		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     aws.String(s.topic),
				Partition: 0,
				Offset:    123,
			},
			Value: s.msg,
		}

		_ = cb(msg)
	}).Return(s.readAllErr)

	mc.On("Close").Return(s.closeErr)
	cg.On("GetConsumer", mock.Anything).Return(mc, s.getConErr)

	rvs, err := resolver.NewResolvers(map[string]interface{}{schema.KeyTypeArticle: new(schema.Article)})
	s.Require().NoError(err)

	pms := &articles.Parameters{
		Consumers: cg,
		Resolvers: rvs,
		Env: &env.Environment{
			MaxParts:                10,
			Partitions:              10,
			ArticleChannelSize:      1,
			ThrottlingMsgsPerSecond: 10,
			Workers:                 1,
		},
		Unmarshaler: s.unmar,
	}

	h := articles.NewHandler(context.Background(), pms, s.topic, schema.KeyTypeArticle)
	rtr := gin.New()
	rtr.POST("/articles", h)
	s.srv = httptest.NewServer(rtr)

	payload, err := json.Marshal(s.reqParams)
	s.Assert().NoError(err)
	req, err := http.NewRequest("POST", s.srv.URL+"/articles", bytes.NewReader(payload))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	client := &http.Client{Timeout: time.Second * 5}
	res, err := client.Do(req)
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Equal(s.status, res.StatusCode)

	if s.status == 200 {
		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			var got schema.Article
			s.Require().NoError(json.Unmarshal([]byte(line), &got))
			s.Equal("TestArticle", got.Name)
		}
	}
}

func TestHandler(t *testing.T) {
	for _, tc := range []*handlerTestSuite{
		{
			status:     200,
			topic:      "test",
			partitions: []int{0},
		},
		{
			status:    422,
			reqParams: &articles.Model{Fields: []string{"version"}},
		},
		{
			status:    422,
			reqParams: &articles.Model{Fields: []string{"not_found"}},
		},
		{
			status: 200,
			reqParams: &articles.Model{Filters: []resolver.RequestFilter{
				{Field: "is_part_of.identifier", Value: "enwiki"},
				{Field: "event.type", Value: "update"},
				{Field: "version.noindex", Value: false},
				{Field: "namespace.identifier", Value: 0},
			}},
		},
		{
			status: 422,
			reqParams: &articles.Model{Filters: []resolver.RequestFilter{
				{Field: "is_part_of", Value: "enwiki"},
			}},
		},
		{
			status: 422,
			reqParams: &articles.Model{Filters: []resolver.RequestFilter{
				{Field: "is_part_of.identifier", Value: 0},
			}},
		},
		{
			status: 422,
			reqParams: &articles.Model{Filters: []resolver.RequestFilter{
				{Field: "categories", Value: "foo"},
			}},
		},
		{
			status: 422,
			reqParams: &articles.Model{Since: time.Now(),
				Parts: []int{0}},
		},
		{
			status: 422,
			reqParams: &articles.Model{
				Parts:             []int{0},
				Offsets:           map[int]int64{0: 0},
				SincePerPartition: map[int]time.Time{0: time.Now()},
			},
		},
	} {
		suite.Run(t, tc)
	}
}
