package env_test

import (
	"encoding/json"
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
	serverPortKey            string
	serverPort               string
	schemaRegistryKey        string
	schemaRegistry           string
	freeTierGroupKey         string
	freeTierGroup            string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5052"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.schemaRegistryKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistry = "localhost:9121"
	s.freeTierGroupKey = "FREE_TIER_GROUP"
	s.freeTierGroup = "group_1"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.schemaRegistryKey, s.schemaRegistry)
	os.Setenv(s.freeTierGroupKey, s.freeTierGroup)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.schemaRegistry, env.SchemaRegistryURL)
	s.Assert().Equal(s.freeTierGroup, env.FreeTierGroup)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
