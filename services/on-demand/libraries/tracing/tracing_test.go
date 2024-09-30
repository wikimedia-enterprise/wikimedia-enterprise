package tracing

import (
	"testing"
	"wikimedia-enterprise/services/on-demand/config/env"

	"github.com/stretchr/testify/suite"
)

type tracingTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *tracingTestSuite) SetupSuite() {
	s.env = new(env.Environment)
}

func (s *tracingTestSuite) TestNew() {
	_, err := NewAPI(s.env)
	s.Assert().Nil(err)
}

func TestTracing(t *testing.T) {
	suite.Run(t, new(tracingTestSuite))
}
