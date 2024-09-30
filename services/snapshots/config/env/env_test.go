package env_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey string
	kafkaBootstrapServers    string
	kafkaCredsKey            string
	kafkaCreds               string
	kafkaAutoOffsetReset     string
	kafkaAutoOffsetResetKey  string
	serverPortKey            string
	serverPort               string
	schemaRegistryKey        string
	schemaRegistry           string
	freeTierGroupKey         string
	freeTierGroup            string
	OTELCollectorAddr        string
	OTELCollectorAddrKey     string
	TracingSamplingRate      float64
	TracingSamplingRateKey   string
	ServiceName              string
	ServiceNameKey           string
	ChunkWorkers             int
	ChunkWorkersKey          string
	LineLimit                int
	LineLimitKey             string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.kafkaAutoOffsetReset = "earliest"
	s.kafkaAutoOffsetResetKey = "KAFKA_AUTO_OFFSET_RESET"
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
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.kafkaAutoOffsetResetKey, s.kafkaAutoOffsetReset)
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.schemaRegistryKey, s.schemaRegistry)
	os.Setenv(s.freeTierGroupKey, s.freeTierGroup)
	os.Setenv(s.OTELCollectorAddrKey, s.OTELCollectorAddr)
	os.Setenv(s.TracingSamplingRateKey, fmt.Sprintf("%f", s.TracingSamplingRate))
	os.Setenv(s.ServiceNameKey, s.ServiceName)
	os.Setenv(s.ChunkWorkersKey, fmt.Sprintf("%d", s.ChunkWorkers))
	os.Setenv(s.LineLimitKey, fmt.Sprintf("%d", s.LineLimit))
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.kafkaAutoOffsetReset, env.KafkaAutoOffsetReset)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.schemaRegistry, env.SchemaRegistryURL)
	s.Assert().Equal(s.freeTierGroup, env.FreeTierGroup)
	s.Assert().Equal(s.ChunkWorkers, env.ChunkWorkers)
	s.Assert().Equal(s.LineLimit, env.LineLimit)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
