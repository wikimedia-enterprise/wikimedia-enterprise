package env_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/services/content-integrity/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	serverPort                           string
	serverPortKey                        string
	prometheusPort                       int
	prometheusPortKey                    string
	redisAddr                            string
	redisAddrKey                         string
	redisPassword                        string
	redisPasswordKey                     string
	kafkaBootstrapServers                string
	kafkaBootstrapServersKey             string
	kafkaConsumerGroupID                 string
	kafkaConsumerGroupIDKey              string
	kafkaCreds                           *env.Credentials
	kafkaCredsKey                        string
	numberOfWorkers                      int
	numberOfWorkersKey                   string
	eventChannelSize                     int
	eventChannelSizeKey                  string
	schemaRegistryCreds                  *env.Credentials
	schemaRegistryCredsKey               string
	schemaRegistryURL                    string
	schemaRegistryURLKey                 string
	topicArticleCreate                   string
	topicArticleCreateKey                string
	topicArticleUpdate                   string
	topicArticleUpdateKey                string
	topicArticleMove                     string
	topicArticleMoveKey                  string
	breakingNewsKeysExpirationKey        string
	breakingNewsKeysExpiration           int
	breakingNewsTemplatesKey             string
	breakingNewsTemplates                env.List
	breakingNewsCreatedHoursKey          string
	breakingNewsCreatedHours             int
	breakingNewsMovedHoursKey            string
	breakingNewsMovedHours               int
	breakingNewsUniqueEditorsKey         string
	breakingNewsUniqueEditors            int
	breakingNewsUniqueEditorsHoursKey    string
	breakingNewsUniqueEditorsHours       int
	breakingNewsEditsHoursKey            string
	breakingNewsEditsHours               int
	breakingNewsEditsKey                 string
	breakingNewsEdits                    int
	breakingNewsTemplatesPrefixKey       string
	breakingNewsTemplatesPrefix          env.List
	breakingNewsTemplatesPrefixIgnoreKey string
	breakingNewsTemplatesPrefixIgnore    env.List
}

func (s *envTestSuite) SetupSuite() {
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5050"

	s.prometheusPortKey = "PROMETHEUS_PORT"
	s.prometheusPort = 12411

	s.redisAddrKey = "REDIS_ADDR"
	s.redisAddr = "redis:6379"

	s.redisPasswordKey = "REDIS_PASSWORD"
	s.redisPassword = "pass"

	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "broker:29092"

	s.kafkaConsumerGroupIDKey = "KAFKA_CONSUMER_GROUP_ID"
	s.kafkaConsumerGroupID = "content-integrity"

	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = &env.Credentials{
		Username: "admin",
		Password: "sql",
	}
	s.numberOfWorkersKey = "NUMBER_OF_WORKERS"
	s.numberOfWorkers = 15

	s.eventChannelSizeKey = "EVENT_CHANNEL_SIZE"
	s.eventChannelSize = 1000000

	s.schemaRegistryCredsKey = "SCHEMA_REGISTRY_CREDS"
	s.schemaRegistryCreds = &env.Credentials{
		Username: "root",
		Password: "root",
	}

	s.schemaRegistryURLKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistryURL = "http://schemaregistry:8085"

	s.topicArticleCreateKey = "TOPIC_ARTICLE_CREATE"
	s.topicArticleCreate = "aws.event-bridge.article-create.v1"

	s.topicArticleUpdateKey = "TOPIC_ARTICLE_UPDATE"
	s.topicArticleUpdate = "aws.event-bridge.article-update.v1"

	s.topicArticleMoveKey = "TOPIC_ARTICLE_MOVE"
	s.topicArticleMove = "aws.event-bridge.article-move.v1"

	s.breakingNewsKeysExpirationKey = "BREAKING_NEWS_KEYS_EXPIRATION"
	s.breakingNewsKeysExpiration = 48

	s.breakingNewsTemplatesKey = "BREAKING_NEWS_TEMPLATES"
	s.breakingNewsTemplates = env.List{"Template:TEST1"}

	s.breakingNewsTemplatesPrefixKey = "BREAKING_NEWS_TEMPLATES_PREFIX"
	s.breakingNewsTemplatesPrefix = env.List{"Template:TEST1"}

	s.breakingNewsTemplatesPrefixIgnoreKey = "BREAKING_NEWS_TEMPLATES_PREFIX_IGNORE"
	s.breakingNewsTemplatesPrefixIgnore = env.List{"Template:TEST1"}

	s.breakingNewsCreatedHoursKey = "BREAKING_NEWS_CREATED_HOURS"
	s.breakingNewsCreatedHours = 24

	s.breakingNewsMovedHoursKey = "BREAKING_NEWS_MOVED_HOURS"
	s.breakingNewsMovedHours = 24

	s.breakingNewsUniqueEditorsKey = "BREAKING_NEWS_UNIQUE_EDITORS"
	s.breakingNewsUniqueEditors = 5

	s.breakingNewsUniqueEditorsHoursKey = "BREAKING_NEWS_UNIQUE_EDITORS_HOURS"
	s.breakingNewsUniqueEditorsHours = 1

	s.breakingNewsEditsHoursKey = "BREAKING_NEWS_EDITS_HOURS"
	s.breakingNewsEditsHours = 1

	s.breakingNewsEditsKey = "BREAKING_NEWS_UNIQUE_EDITS"
	s.breakingNewsEdits = 10
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.prometheusPortKey, strconv.Itoa(s.prometheusPort))
	os.Setenv(s.redisAddrKey, s.redisAddr)
	os.Setenv(s.redisPasswordKey, s.redisPassword)
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaConsumerGroupIDKey, s.kafkaConsumerGroupID)
	os.Setenv(s.numberOfWorkersKey, strconv.Itoa(s.numberOfWorkers))
	os.Setenv(s.eventChannelSizeKey, strconv.Itoa(s.eventChannelSize))
	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
	os.Setenv(s.topicArticleCreateKey, s.topicArticleCreate)
	os.Setenv(s.topicArticleUpdateKey, s.topicArticleUpdate)
	os.Setenv(s.topicArticleMoveKey, s.topicArticleMove)
	os.Setenv(s.breakingNewsKeysExpirationKey, strconv.Itoa(s.breakingNewsKeysExpiration))
	os.Setenv(s.breakingNewsCreatedHoursKey, strconv.Itoa(s.breakingNewsCreatedHours))
	os.Setenv(s.breakingNewsMovedHoursKey, strconv.Itoa(s.breakingNewsMovedHours))
	os.Setenv(s.breakingNewsUniqueEditorsHoursKey, strconv.Itoa(s.breakingNewsUniqueEditorsHours))
	os.Setenv(s.breakingNewsUniqueEditorsKey, strconv.Itoa(s.breakingNewsUniqueEditors))
	os.Setenv(s.breakingNewsEditsHoursKey, strconv.Itoa(s.breakingNewsEditsHours))
	os.Setenv(s.breakingNewsEditsKey, strconv.Itoa(s.breakingNewsEdits))

	kfc, err := json.Marshal(s.kafkaCreds)
	s.Assert().NoError(err)
	os.Setenv(s.kafkaCredsKey, string(kfc))

	src, err := json.Marshal(s.schemaRegistryCreds)
	s.Assert().NoError(err)
	os.Setenv(s.schemaRegistryCredsKey, string(src))

	tmp, err := json.Marshal(s.breakingNewsTemplates)
	s.Assert().NoError(err)
	os.Setenv(s.breakingNewsTemplatesKey, string(tmp))

	pfx, err := json.Marshal(s.breakingNewsTemplates)
	s.Assert().NoError(err)
	os.Setenv(s.breakingNewsTemplatesPrefixKey, string(pfx))

	ign, err := json.Marshal(s.breakingNewsTemplates)
	s.Assert().NoError(err)
	os.Setenv(s.breakingNewsTemplatesPrefixIgnoreKey, string(ign))

}

func (s *envTestSuite) TestNew() {
	env, err := env.New()

	s.Assert().NoError(err)
	s.Assert().NotNil(env)

	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.prometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.redisAddr, env.RedisAddr)
	s.Assert().Equal(s.redisPassword, env.RedisPassword)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.kafkaConsumerGroupID, env.KafkaConsumerGroupID)
	s.Assert().Equal(s.numberOfWorkers, env.NumberOfWorkers)
	s.Assert().Equal(s.eventChannelSize, env.EventChannelSize)
	s.Assert().Equal(s.schemaRegistryURL, env.SchemaRegistryURL)
	s.Assert().Equal(s.topicArticleCreate, env.TopicArticleCreate)
	s.Assert().Equal(s.topicArticleUpdate, env.TopicArticleUpdate)
	s.Assert().Equal(s.topicArticleMove, env.TopicArticleMove)
	s.Assert().Equal(s.kafkaCreds, env.KafkaCreds)
	s.Assert().Equal(s.schemaRegistryCreds, env.SchemaRegistryCreds)

	s.Assert().Equal(s.breakingNewsKeysExpiration, env.BreakingNewsKeysExpiration)
	s.Assert().Equal(s.breakingNewsTemplates, env.BreakingNewsMandatoryTemplates)
	s.Assert().Equal(s.breakingNewsTemplatesPrefix, env.BreakingNewsTemplatesPrefix)
	s.Assert().Equal(s.breakingNewsTemplatesPrefixIgnore, env.BreakingNewsTemplatesPrefixIgnore)
	s.Assert().Equal(s.breakingNewsCreatedHours, env.BreakingNewsCreatedHours)
	s.Assert().Equal(s.breakingNewsMovedHours, env.BreakingNewsMovedHours)
	s.Assert().Equal(s.breakingNewsUniqueEditorsHours, env.BreakingNewsUniqueEditorsHours)
	s.Assert().Equal(s.breakingNewsUniqueEditors, env.BreakingNewsUniqueEditors)
	s.Assert().Equal(s.breakingNewsEditsHours, env.BreakingNewsEditsHours)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
