package stream_test

import (
	"testing"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/libraries/stream"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/suite"
)

type streamTestSuite struct {
	suite.Suite
	url   string
	creds *env.Credentials
	env   *env.Environment
}

func (s *streamTestSuite) SetupSuite() {
	s.env = new(env.Environment)

	if s.creds != nil {
		s.env.SchemaRegistryCreds = s.creds
	}

	s.env.SchemaRegistryURL = s.url
}

func (s *streamTestSuite) TestNew() {
	s.Assert().NotNil(stream.New(s.env, new(kafka.Producer)))
}

func TestNew(t *testing.T) {
	for _, testCase := range []*streamTestSuite{
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
}
