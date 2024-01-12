package kafka_test

import (
	"testing"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/libraries/kafka"

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

func (s *kafkaTestSuite) TestNewConsumer() {
	cnn, err := kafka.NewConsumer(s.env)
	s.Assert().NoError(err)
	s.Assert().NotNil(cnn)
}

func (s *kafkaTestSuite) TestNewProducer() {
	prd, err := kafka.NewProducer(s.env)
	s.Assert().NoError(err)
	s.Assert().NotNil(prd)
}

func TestKafka(t *testing.T) {
	suite.Run(t, new(kafkaTestSuite))
}
