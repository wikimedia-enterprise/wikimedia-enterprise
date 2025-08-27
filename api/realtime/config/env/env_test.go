package env_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/api/realtime/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	serverPortKey                     string
	serverPort                        string
	serverModeKey                     string
	serverMode                        string
	ksqlURLkey                        string
	ksqlURL                           string
	KSQLUsernameKey                   string
	KSQLUsername                      string
	KSQLPasswordKey                   string
	KSQLPassword                      string
	awsRegionKey                      string
	awsRegion                         string
	awsIDKey                          string
	awsID                             string
	awsKeyKey                         string
	awsKey                            string
	clientIDKey                       string
	clientID                          string
	clientSecretKey                   string
	clientSecret                      string
	redisAddrKey                      string
	redisAddr                         string
	redisPasswordKey                  string
	redisPassword                     string
	prometheusPortKey                 string
	prometheusPort                    int
	partitionsKey                     string
	partitions                        int
	maxPartsKey                       string
	maxParts                          int
	workersKey                        string
	workers                           int
	articlesStreamKey                 string
	articlesStream                    string
	IPAllowList                       string
	IPAllowListKey                    string
	KSQLHealthCheckTimeoutSecondsKey  string
	KSQLHealthCheckIntervalSeconds    int
	KSQLHealthCheckTimeoutSeconds     int
	KSQLHealthCheckIntervalSecondsKey string
	redisHealthCheckTimeoutSecondsKey string
	redisHealthCheckTimeoutSeconds    int
	articlesTopicKey                  string
	articlesTopic                     string
	structuredTopicKey                string
	structuredTopic                   string
	entityTopicKey                    string
	entityTopic                       string
	articleChannelSizeKey             string
	articleChannelSize                int
	kafkaBootstrapServersKey          string
	kafkaBootstrapServers             string
	kafkaCredsKey                     string
	kafkaCreds                        string
	schemaRegistryKey                 string
	schemaRegistry                    string
	AccessModel                       string
	AccessModelKey                    string
	AccessPolicy                      string
	AccessPolicyKey                   string
}

func (s *envTestSuite) SetupSuite() {
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5052"
	s.serverModeKey = "SERVER_MODE"
	s.serverMode = "debug"
	s.ksqlURLkey = "KSQL_URL"
	s.ksqlURL = "localhost:2020"
	s.KSQLUsernameKey = "KSQL_USERNAME"
	s.KSQLUsername = "foo"
	s.KSQLPasswordKey = "KSQL_PASSWORD"
	s.KSQLPassword = "bar"
	s.awsRegionKey = "AWS_REGION"
	s.awsRegion = "region-1"
	s.awsIDKey = "AWS_ID"
	s.awsID = "foo"
	s.awsKeyKey = "AWS_KEY"
	s.awsKey = "bar"
	s.clientIDKey = "COGNITO_CLIENT_ID"
	s.clientID = "client_id"
	s.clientSecretKey = "COGNITO_CLIENT_SECRET"
	s.clientSecret = "secret"
	s.redisAddrKey = "REDIS_ADDR"
	s.redisAddr = "redis:2020"
	s.redisPasswordKey = "REDIS_PASSWORD"
	s.redisPassword = "secret"
	s.prometheusPortKey = "PROMETHEUS_PORT"
	s.prometheusPort = 101
	s.partitionsKey = "PARTITIONS"
	s.partitions = 100
	s.maxPartsKey = "MAX_PARTS"
	s.maxParts = 10
	s.workersKey = "WORKERS"
	s.workers = 10
	s.articlesStreamKey = "ARTICLES_STREAM"
	s.articlesStream = "articles_str"
	s.IPAllowListKey = "IP_ALLOW_LIST"
	s.IPAllowList = "[{\"ip_range\": {\"start\": \"192.168.0.1\", \"end\": \"192.168.0.10\"}, \"user\": {\"username\": \"user1\", \"groups\": [\"group_1\"]}}, {\"ip_range\": {\"start\": \"192.168.1.1\", \"end\": \"192.168.1.10\"}, \"user\": {\"username\": \"user2\", \"groups\": [\"group_2\"]}}]"
	s.KSQLHealthCheckTimeoutSeconds = 5
	s.KSQLHealthCheckTimeoutSecondsKey = "KSQL_HEALTH_TIMEOUT"
	s.KSQLHealthCheckIntervalSeconds = 10
	s.KSQLHealthCheckIntervalSecondsKey = "KSQL_HEALTH_INTERVAL"
	s.redisHealthCheckTimeoutSecondsKey = "REDIS_HEALTH_TIMEOUT"
	s.redisHealthCheckTimeoutSeconds = 5
	s.articleChannelSizeKey = "ARTICLE_CHANNEL_SIZE"
	s.articleChannelSize = 100
	s.articlesTopicKey = "ARTICLES_TOPIC"
	s.articlesTopic = "articles"
	s.entityTopicKey = "ENTITY_TOPIC"
	s.entityTopic = "entities"
	s.structuredTopicKey = "STRUCTURED_TOPIC"
	s.structuredTopic = "str"
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.schemaRegistryKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistry = "localhost:9121"
	s.AccessModelKey = "ACCESS_MODEL"
	s.AccessModel = "model content"
	s.AccessPolicyKey = "ACCESS_POLICY"
	s.AccessPolicy = "policy content"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.serverModeKey, s.serverMode)
	os.Setenv(s.ksqlURLkey, s.ksqlURL)
	os.Setenv(s.KSQLUsernameKey, s.KSQLUsername)
	os.Setenv(s.KSQLPasswordKey, s.KSQLPassword)
	os.Setenv(s.awsRegionKey, s.awsRegion)
	os.Setenv(s.awsIDKey, s.awsID)
	os.Setenv(s.awsKeyKey, s.awsKey)
	os.Setenv(s.clientIDKey, s.clientID)
	os.Setenv(s.clientSecretKey, s.clientSecret)
	os.Setenv(s.redisAddrKey, s.redisAddr)
	os.Setenv(s.redisPasswordKey, s.redisPassword)
	os.Setenv(s.prometheusPortKey, strconv.Itoa(s.prometheusPort))
	os.Setenv(s.partitionsKey, strconv.Itoa(s.partitions))
	os.Setenv(s.maxPartsKey, strconv.Itoa(s.maxParts))
	os.Setenv(s.workersKey, strconv.Itoa(s.workers))
	os.Setenv(s.articlesStreamKey, s.articlesStream)
	os.Setenv(s.IPAllowListKey, s.IPAllowList)
	os.Setenv(s.KSQLHealthCheckTimeoutSecondsKey, strconv.Itoa(s.KSQLHealthCheckTimeoutSeconds))
	os.Setenv(s.KSQLHealthCheckIntervalSecondsKey, strconv.Itoa(s.KSQLHealthCheckIntervalSeconds))
	os.Setenv(s.redisHealthCheckTimeoutSecondsKey, strconv.Itoa(s.redisHealthCheckTimeoutSeconds))
	os.Setenv(s.articleChannelSizeKey, strconv.Itoa(s.articleChannelSize))
	os.Setenv(s.articlesTopicKey, s.articlesTopic)
	os.Setenv(s.entityTopicKey, s.entityTopic)
	os.Setenv(s.structuredTopicKey, s.structuredTopic)
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.schemaRegistryKey, s.schemaRegistry)
	os.Setenv(s.AccessModelKey, s.AccessModel)
	os.Setenv(s.AccessPolicyKey, s.AccessPolicy)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()

	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.serverMode, env.ServerMode)
	s.Assert().Equal(s.ksqlURL, env.KSQLURL)
	s.Assert().Equal(s.KSQLUsername, env.KSQLUsername)
	s.Assert().Equal(s.KSQLPassword, env.KSQLPassword)
	s.Assert().Equal(s.awsRegion, env.AWSRegion)
	s.Assert().Equal(s.awsID, env.AWSID)
	s.Assert().Equal(s.awsKey, env.AWSKey)
	s.Assert().Equal(s.clientID, env.CognitoClientID)
	s.Assert().Equal(s.clientSecret, env.CognitoClientSecret)
	s.Assert().Equal(s.redisAddr, env.RedisAddr)
	s.Assert().Equal(s.redisPassword, env.RedisPassword)
	s.Assert().Equal(s.prometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.partitions, env.Partitions)
	s.Assert().Equal(s.maxParts, env.MaxParts)
	s.Assert().Equal(s.workers, env.Workers)
	s.Assert().Equal(s.articlesStream, env.ArticlesStream)
	s.Assert().NotNil(env.IPAllowList)
	s.Assert().Equal(s.articleChannelSize, env.ArticleChannelSize)
	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))

	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.schemaRegistry, env.SchemaRegistryURL)
	s.Assert().NotNil(env.AccessModel.Path)
	s.Assert().NotNil(env.AccessPolicy.Path)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
