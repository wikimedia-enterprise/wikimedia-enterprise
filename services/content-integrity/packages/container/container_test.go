package container_test

import (
	"os"
	"testing"
	"wikimedia-enterprise/services/content-integrity/config/env"
	"wikimedia-enterprise/services/content-integrity/packages/container"

	"github.com/stretchr/testify/suite"
)

type containerTestSuite struct {
	suite.Suite
	redisAddr                string
	redisAddrKey             string
	kafkaBootstrapServers    string
	kafkaBootstrapServersKey string
	schemaRegistryURL        string
	schemaRegistryURLKey     string
}

func (s *containerTestSuite) SetupSuite() {
	s.redisAddrKey = "REDIS_ADDR"
	s.redisAddr = "redis:123"

	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "botstrap:1234"

	s.schemaRegistryURLKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistryURL = "http://schemaregistry:8085"
}

func (s *containerTestSuite) SetupTest() {
	os.Setenv(s.redisAddrKey, s.redisAddr)
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
}

func (s *containerTestSuite) TestNew() {
	cnt, err := container.New()

	s.Assert().NoError(err)
	s.Assert().NotNil(cnt)
	s.Assert().NotPanics(func() {
		s.Assert().NoError(cnt.Invoke(func(env *env.Environment) {
			s.Assert().NotNil(env)
		}))
	})
}

func TestContainer(t *testing.T) {
	suite.Run(t, new(containerTestSuite))
}
