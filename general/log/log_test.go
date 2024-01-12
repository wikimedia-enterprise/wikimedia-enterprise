package log

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type logTestSuite struct {
	suite.Suite
}

func (s *logTestSuite) TestNew() {
	lgr, err := New()

	s.Assert().NotNil(lgr)
	s.Assert().NoError(err)
}

func (s *logTestSuite) TestLogger() {
	s.Assert().NotNil(logger)
}

func (s *logTestSuite) TestGetZap() {
	s.Assert().NotNil(GetZap())
}

func TestLogger(t *testing.T) {
	suite.Run(t, new(logTestSuite))
}
