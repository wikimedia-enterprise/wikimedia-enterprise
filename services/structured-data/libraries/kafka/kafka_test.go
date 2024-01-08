package kafka_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/libraries/kafka"

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

func (s *kafkaTestSuite) TestNewConsumer() {
	conn, err := kafka.NewConsumer(s.env)
	s.Assert().NoError(err)
	s.Assert().NotNil(conn)
}

func TestKafka(t *testing.T) {
	suite.Run(t, new(kafkaTestSuite))
}
