package httputil

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type recoveryTestSuite struct {
	suite.Suite
}

func (s *recoveryTestSuite) TestRecovery() {
	s.Assert().NotNil(Recovery(zap.NewNop()))
}

func TestRecovery(t *testing.T) {
	suite.Run(t, new(recoveryTestSuite))
}
