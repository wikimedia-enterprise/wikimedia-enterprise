package env_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/services/structured-contents/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey   string
	kafkaBootstrapServers      string
	kafkaConsumerGroupID       string
	kafkaConsumerGroupIDKey    string
	kafkaCreds                 string
	kafkaCredsKey              string
	kafkaAutoOffsetReset       string
	kafkaAutoOffsetResetKey    string
	kafkaMaxPollIntervalKey    string
	kafkaMaxPollInterval       int
	schemaRegistryURLKey       string
	schemaRegistryURL          string
	schemaRegistryCredsKey     string
	schemaRegistryCreds        string
	inputTopics                env.List
	inputTopicsKey             string
	outputCompactedTopics      string
	outputCompactedTopicsKey   string
	outputNoncompactedTopic    string
	outputNoncompactedTopicKey string
	numberOfWorkersKey         string
	numberOfWorkers            int
	eventChannelSizeKey        string
	eventChannelSize           int
	tracingGRPCHostKey         string
	tracingGRPCHost            string
	tracingSamplingRate        float64
	tracingSamplingRateKey     string
	serviceName                string
	serviceNameKey             string
	prometheusPort             int
	prometheusPortKey          string
	deadLetterTopic            string
	deadLetterTopicKey         string
	maxFailCount               string
	maxFailCountKey            string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.kafkaConsumerGroupIDKey = "KAFKA_CONSUMER_GROUP_ID"
	s.kafkaConsumerGroupID = "unique"
	s.kafkaAutoOffsetReset = "earliest"
	s.kafkaAutoOffsetResetKey = "KAFKA_AUTO_OFFSET_RESET"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.kafkaMaxPollIntervalKey = "KAFKA_MAX_POLL_INTERVAL"
	s.kafkaMaxPollInterval = 0
	s.outputCompactedTopics = `{"version":["v1"],"service_name":"some-service","location":"local"}`
	s.outputCompactedTopicsKey = "OUTPUT_COMPACTED_TOPICS"
	s.outputNoncompactedTopic = "aws.structured-contents.articles.v1"
	s.outputNoncompactedTopicKey = "OUTPUT_NONCOMPACTED_TOPIC"
	s.schemaRegistryURLKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistryURL = "localhost:2020"
	s.schemaRegistryCredsKey = "SCHEMA_REGISTRY_CREDS"
	s.schemaRegistryCreds = `{"username":"reg_admin","password":"654321"}`
	s.numberOfWorkersKey = "NUMBER_OF_WORKERS"
	s.numberOfWorkers = 10
	s.eventChannelSizeKey = "EVENT_CHANNEL_SIZE"
	s.eventChannelSize = 100000
	s.inputTopics = env.List{"aws.structured-data.articles.v1"}
	s.inputTopicsKey = "INPUT_TOPICS"
	s.serviceName = "structured-contents.service"
	s.serviceNameKey = "SERVICE_NAME"
	s.tracingGRPCHost = "collector"
	s.tracingGRPCHostKey = "OTEL_COLLECTOR_ADDR"
	s.tracingSamplingRate = 0.1
	s.tracingSamplingRateKey = "TRACING_SAMPLING_RATE"
	s.prometheusPortKey = "PROMETHEUS_PORT"
	s.prometheusPort = 12411
	s.serviceName = "service.x"
	s.serviceNameKey = "SERVICE_NAME"
	s.deadLetterTopic = "aws.structured-contents.bulk-dead-letter.v1"
	s.deadLetterTopicKey = "DEAD_LETTER_TOPIC"
	s.maxFailCountKey = "MAX_FAIL_COUNT"
	s.maxFailCount = "1"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaConsumerGroupIDKey, s.kafkaConsumerGroupID)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.kafkaAutoOffsetResetKey, s.kafkaAutoOffsetReset)

	inp, err := json.Marshal(s.inputTopics)
	s.Assert().NoError(err)
	os.Setenv(s.inputTopicsKey, string(inp))

	os.Setenv(s.outputNoncompactedTopicKey, s.outputNoncompactedTopic)
	os.Setenv(s.numberOfWorkersKey, strconv.Itoa(s.numberOfWorkers))
	os.Setenv(s.eventChannelSizeKey, strconv.Itoa(s.eventChannelSize))
	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
	os.Setenv(s.schemaRegistryCredsKey, s.schemaRegistryCreds)
	os.Setenv(s.serviceNameKey, s.serviceName)
	os.Setenv(s.tracingGRPCHostKey, s.tracingGRPCHost)
	os.Setenv(s.prometheusPortKey, strconv.Itoa(s.prometheusPort))
	os.Setenv(s.tracingSamplingRateKey, strconv.FormatFloat(s.tracingSamplingRate, 'f', -1, 64))
	os.Setenv(s.deadLetterTopicKey, s.deadLetterTopic)
	os.Setenv(s.maxFailCountKey, s.maxFailCount)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal("structured-contents", env.OutputCompactedTopics.ServiceName)

	os.Setenv(s.outputCompactedTopicsKey, s.outputCompactedTopics)
	s.Assert().Equal(s.inputTopics, env.InputTopics)
	s.Assert().Equal(s.outputNoncompactedTopic, env.OutputNoncompactedTopic)

	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.kafkaAutoOffsetReset, env.KafkaAutoOffsetReset)
	s.Assert().Equal(s.kafkaMaxPollInterval, env.KafkaMaxPollInterval)

	s.Assert().Equal(s.kafkaConsumerGroupID, env.KafkaConsumerGroupID)
	s.Assert().Equal(s.schemaRegistryURL, env.SchemaRegistryURL)
	s.Assert().Equal(s.numberOfWorkers, env.NumberOfWorkers)
	s.Assert().Equal(s.eventChannelSize, env.EventChannelSize)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))

	creds, err = json.Marshal(env.SchemaRegistryCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.schemaRegistryCreds, string(creds))

	s.Assert().Equal(s.serviceName, env.ServiceName)
	s.Assert().Equal(s.tracingGRPCHost, env.TracingGRPCHost)
	s.Assert().Equal(s.tracingSamplingRate, env.TracingSamplingRate)
	s.Assert().Equal(s.prometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.deadLetterTopic, env.DeadLetterTopic)
	max, err := strconv.Atoi(s.maxFailCount)
	s.Assert().NoError(err)
	s.Assert().Equal(max, env.MaxFailCount)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
