package prometheus

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// prometheusTestSuite defines the test suite for the prometheus package
type prometheusTestSuite struct {
	suite.Suite
}

// TestNewAPI tests the NewAPI function
func (s *prometheusTestSuite) TestNewAPI() {
	metrics := NewAPI()
	s.Assert().NotNil(metrics)
}

// TestPrometheus is the entry point for the test suite
func TestPrometheus(t *testing.T) {
	suite.Run(t, new(prometheusTestSuite))
}
