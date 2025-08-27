package env_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey            string
	kafkaBootstrapServers               string
	kafkaCredsKey                       string
	kafkaCreds                          string
	kafkaAutoOffsetReset                string
	kafkaAutoOffsetResetKey             string
	kafkaDebugOptions                   string
	kafkaDebugOptionsKey                string
	kafkaLogWatermark                   bool
	kafkaLogWatermarkKey                string
	kafkaUseWatermark                   bool
	kafkaUseWatermarkKey                string
	serverPortKey                       string
	serverPort                          string
	schemaRegistryKey                   string
	schemaRegistry                      string
	freeTierGroupKey                    string
	freeTierGroup                       string
	OTELCollectorAddr                   string
	OTELCollectorAddrKey                string
	TracingSamplingRate                 float64
	TracingSamplingRateKey              string
	ServiceName                         string
	ServiceNameKey                      string
	ChunkWorkers                        int
	ChunkWorkersKey                     string
	LineLimit                           int
	LineLimitKey                        string
	CommonsAggregationWorkercount       int
	CommonsAggregationWorkercountKey    string
	AWSBucket                           string
	AWSBucketKey                        string
	AWSBucketCommons                    string
	AWSBucketCommonsKey                 string
	LogInterval                         int
	LogIntervalKey                      string
	CommonsAggregationKeyChannelSize    int
	CommonsAggregationKeyChannelSizeKey string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.kafkaAutoOffsetReset = "earliest"
	s.kafkaAutoOffsetResetKey = "KAFKA_AUTO_OFFSET_RESET"
	s.kafkaDebugOptions = "fetch,topic"
	s.kafkaDebugOptionsKey = "KAFKA_DEBUG_OPTIONS"
	s.kafkaLogWatermark = true
	s.kafkaLogWatermarkKey = "KAFKA_LOG_WATERMARK"
	s.kafkaUseWatermark = true
	s.kafkaUseWatermarkKey = "KAFKA_USE_WATERMARK"
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5052"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.schemaRegistryKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistry = "localhost:9121"
	s.freeTierGroupKey = "FREE_TIER_GROUP"
	s.freeTierGroup = "group_1"
	s.OTELCollectorAddrKey = "OTEL_COLLECTOR_ADDR"
	s.OTELCollectorAddr = "collector:4317"
	s.TracingSamplingRateKey = "TRACING_SAMPLING_RATE"
	s.TracingSamplingRate = 0.1
	s.ServiceNameKey = "SERVICE_NAME"
	s.ServiceName = "snapshots.service"
	s.ChunkWorkersKey = "CHUNK_WORKERS"
	s.ChunkWorkers = 3
	s.LineLimit = 5000
	s.LineLimitKey = "LINE_LIMIT"
	s.CommonsAggregationWorkercount = 10
	s.CommonsAggregationWorkercountKey = "COMMONS_AGGREGATION_WORKER_COUNT"
	s.CommonsAggregationKeyChannelSize = 10
	s.CommonsAggregationKeyChannelSizeKey = "COMMONS_AGGREGATION_KEY_CHANNEL_SIZE"
	s.LogInterval = 10
	s.LogIntervalKey = "LOG_INTERVAL"
	s.AWSBucketKey = "AWS_BUCKET"
	s.AWSBucket = "wme-primary"
	s.AWSBucketCommonsKey = "AWS_BUCKET_COMMONS"
	s.AWSBucketCommons = "wme-data"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.kafkaAutoOffsetResetKey, s.kafkaAutoOffsetReset)
	os.Setenv(s.kafkaDebugOptionsKey, s.kafkaDebugOptions)
	os.Setenv(s.kafkaLogWatermarkKey, strconv.FormatBool(s.kafkaLogWatermark))
	os.Setenv(s.kafkaUseWatermarkKey, strconv.FormatBool(s.kafkaUseWatermark))
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.schemaRegistryKey, s.schemaRegistry)
	os.Setenv(s.freeTierGroupKey, s.freeTierGroup)
	os.Setenv(s.OTELCollectorAddrKey, s.OTELCollectorAddr)
	os.Setenv(s.TracingSamplingRateKey, fmt.Sprintf("%f", s.TracingSamplingRate))
	os.Setenv(s.ServiceNameKey, s.ServiceName)
	os.Setenv(s.ChunkWorkersKey, fmt.Sprintf("%d", s.ChunkWorkers))
	os.Setenv(s.LineLimitKey, fmt.Sprintf("%d", s.LineLimit))
	os.Setenv(s.CommonsAggregationWorkercountKey, fmt.Sprintf("%d", s.CommonsAggregationWorkercount))
	os.Setenv(s.AWSBucketKey, s.AWSBucket)
	os.Setenv(s.AWSBucketCommonsKey, s.AWSBucketCommons)
	os.Setenv(s.LogIntervalKey, fmt.Sprintf("%d", s.LogInterval))
	os.Setenv(s.CommonsAggregationKeyChannelSizeKey, fmt.Sprintf("%d", s.CommonsAggregationKeyChannelSize))
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.kafkaAutoOffsetReset, env.KafkaAutoOffsetReset)
	s.Assert().Equal(s.kafkaDebugOptions, env.KafkaDebugOptions)
	s.Assert().Equal(s.kafkaLogWatermark, env.KafkaLogWatermark)
	s.Assert().Equal(s.kafkaUseWatermark, env.KafkaUseWatermark)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.schemaRegistry, env.SchemaRegistryURL)
	s.Assert().Equal(s.freeTierGroup, env.FreeTierGroup)
	s.Assert().Equal(s.ChunkWorkers, env.ChunkWorkers)
	s.Assert().Equal(s.LineLimit, env.LineLimit)
	s.Assert().Equal(s.LogInterval, env.LogInterval)
	s.Assert().Equal(s.CommonsAggregationWorkercount, env.CommonsAggregationWorkercount)
	s.Assert().Equal(s.CommonsAggregationKeyChannelSize, env.CommonsAggregationKeyChannelSize)
	s.Assert().Equal(s.AWSBucket, env.AWSBucket)
	s.Assert().Equal(s.AWSBucketCommons, env.AWSBucketCommons)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
