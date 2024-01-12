package stream_test

import (
	"testing"
	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/libraries/stream"

	"github.com/stretchr/testify/suite"
)

type streamTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *streamTestSuite) TestNew() {
	s.Assert().NotNil(stream.New(s.env))
}

func TestNew(t *testing.T) {
	for _, testCase := range []*streamTestSuite{
		{
			env: &env.Environment{
				SchemaRegistryURL: "localhost:8080",
			},
		},
		{
			env: &env.Environment{
				SchemaRegistryURL: "localhost:8080",
				SchemaRegistryCreds: &env.Credentials{
					Username: "admin",
					Password: "sql",
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
