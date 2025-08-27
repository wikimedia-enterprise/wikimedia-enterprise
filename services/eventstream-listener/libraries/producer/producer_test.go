package producer_test

import (
	"testing"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/libraries/producer"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type producerTestSuite struct {
	suite.Suite
	url   string
	creds *env.Credentials
	env   *env.Environment
}

func (s *producerTestSuite) SetupSuite() {
	s.env = new(env.Environment)

	if s.creds != nil {
		s.env.SchemaRegistryCreds = s.creds
	}

	s.env.SchemaRegistryURL = s.url
}

func (s *producerTestSuite) TestNew() {
	s.Assert().NotNil(producer.New(s.env, new(kafka.Producer)))
}

func TestNew(t *testing.T) {
	for _, testCase := range []*producerTestSuite{
		{
			url: "localhost:8080",
		},
		{
			url: "localhost:8081",
			creds: &env.Credentials{
				Username: "admin",
				Password: "sql",
			},
		},
	} {
		suite.Run(t, testCase)
	}

	assert.NotNil(t, producer.New(new(env.Environment), new(kafka.Producer)))

}
