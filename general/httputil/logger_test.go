package httputil

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type loggerTestSuite struct {
	suite.Suite
}

func (s *loggerTestSuite) TestLogger() {
	s.Assert().NotNil(Logger(zap.NewNop()))
}

func TestLogger(t *testing.T) {
	suite.Run(t, new(loggerTestSuite))
}
