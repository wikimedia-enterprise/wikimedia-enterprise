package env_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"wikimedia-enterprise/services/event-bridge/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey        string
	kafkaCredsKey                   string
	redisAddrKey                    string
	redisPasswordKey                string
	schemaRegistryCredsKey          string
	schemaRegistryURLKey            string
	topicArticleDeleteKey           string
	topicArticleUpdateKey           string
	topicArticleCreateKey           string
	topicArticleMoveKey             string
	topicArticleVisibilityChangeKey string
	OTELCollectorAddrKey            string
	TracingSamplingRateKey          string
	ServiceNameKey                  string
	kafkaBootstrapServers           string
	kafkaCreds                      string
	redisAddr                       string
	redisPassword                   string
	schemaRegistryURL               string
	schemaRegistryCreds             string
	topicArticleDelete              string
	topicArticleUpdate              string
	topicArticleCreate              string
	topicArticleMove                string
	topicArticleVisibilityChange    string
	OTELCollectorAddr               string
	TracingSamplingRate             float64
	ServiceName                     string
	OutputTopics                    string
	OutputTopicsKey                 string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.redisAddrKey = "REDIS_ADDR"
	s.kafkaBootstrapServers = "localhost:8001"
	s.redisAddr = "localhost:2021"
	s.topicArticleDeleteKey = "TOPIC_ARTICLE_DELETE"
	s.topicArticleDelete = "local.event-bridge.article-delete.v1"
	s.topicArticleUpdateKey = "TOPIC_ARTICLE_UPDATE"
	s.topicArticleUpdate = "local.event-bridge.article-update.v1"
	s.topicArticleCreateKey = "TOPIC_ARTICLE_CREATE"
	s.topicArticleCreate = "local.event-bridge.article-create.v1"
	s.topicArticleMoveKey = "TOPIC_ARTICLE_MOVE"
	s.topicArticleMove = "local.event-bridge.article-move.v1"
	s.topicArticleVisibilityChangeKey = "TOPIC_ARTICLE_VISIBILITY_CHANGE"
	s.topicArticleVisibilityChange = "local.event-bridge.article-visibility-change.v1"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.redisPasswordKey = "REDIS_PASSWORD"
	s.redisPassword = "12345"
	s.schemaRegistryURLKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistryURL = "localhost:300"
	s.schemaRegistryCredsKey = "SCHEMA_REGISTRY_CREDS"
	s.schemaRegistryCreds = `{"username":"reg_admin","password":"654321"}`
	s.OTELCollectorAddrKey = "OTEL_COLLECTOR_ADDR"
	s.OTELCollectorAddr = "collector:4317"
	s.TracingSamplingRateKey = "TRACING_SAMPLING_RATE"
	s.TracingSamplingRate = 0.1
	s.ServiceNameKey = "SERVICE_NAME"
	s.ServiceName = "event-bridge.service"
	s.OutputTopics = "{\"version\":[\"v1\"],\"service_name\":\"event-bridge\",\"location\":\"local\"}"
	s.OutputTopicsKey = "OUTPUT_TOPICS"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.redisAddrKey, s.redisAddr)
	os.Setenv(s.topicArticleDeleteKey, s.topicArticleDelete)
	os.Setenv(s.topicArticleUpdateKey, s.topicArticleUpdate)
	os.Setenv(s.topicArticleCreateKey, s.topicArticleCreate)
	os.Setenv(s.topicArticleMoveKey, s.topicArticleMove)
	os.Setenv(s.topicArticleVisibilityChangeKey, s.topicArticleVisibilityChange)
	os.Setenv(s.redisPasswordKey, s.redisPassword)
	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
	os.Setenv(s.schemaRegistryCredsKey, s.schemaRegistryCreds)
	os.Setenv(s.OTELCollectorAddrKey, s.OTELCollectorAddr)
	os.Setenv(s.TracingSamplingRateKey, fmt.Sprintf("%f", s.TracingSamplingRate))
	os.Setenv(s.ServiceNameKey, s.ServiceName)
	os.Setenv(s.OutputTopicsKey, s.OutputTopics)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.redisAddr, env.RedisAddr)
	s.Assert().Equal(s.topicArticleDelete, env.TopicArticleDelete)
	s.Assert().Equal(s.topicArticleUpdate, env.TopicArticleUpdate)
	s.Assert().Equal(s.topicArticleCreate, env.TopicArticleCreate)
	s.Assert().Equal(s.topicArticleMove, env.TopicArticleMove)
	s.Assert().Equal(s.topicArticleVisibilityChange, env.TopicArticleVisibilityChange)
	s.Assert().Equal(s.redisPassword, env.RedisPassword)
	s.Assert().Equal(s.schemaRegistryURL, env.SchemaRegistryURL)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))

	creds, err = json.Marshal(env.SchemaRegistryCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.schemaRegistryCreds, string(creds))

	s.Assert().Equal(s.OTELCollectorAddr, env.OTELCollectorAddr)
	s.Assert().Equal(s.TracingSamplingRate, env.TracingSamplingRate)
	s.Assert().Equal(s.ServiceName, env.ServiceName)

	tps, err := json.Marshal(env.OutputTopics)
	s.Assert().NoError(err)
	s.Assert().Equal(s.OutputTopics, string(tps))
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
