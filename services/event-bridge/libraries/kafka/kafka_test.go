package kafka_test

import (
	"testing"
	"wikimedia-enterprise/services/event-bridge/config/env"
	"wikimedia-enterprise/services/event-bridge/libraries/kafka"

	"github.com/stretchr/testify/suite"
)

type kafkaTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *kafkaTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.KafkaBootstrapServers = "localhost"
	s.env.KafkaCreds = &env.Credentials{
		Username: "admin",
		Password: "sql",
	}
}

func (s *kafkaTestSuite) TestNewProducer() {
	prod, err := kafka.NewProducer(s.env)
	s.Assert().NoError(err)
	s.Assert().NotNil(prod)
}

func TestKafka(t *testing.T) {
	suite.Run(t, new(kafkaTestSuite))
}
