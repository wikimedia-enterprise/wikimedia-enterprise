package prometheus_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-data/libraries/prometheus"

	"github.com/stretchr/testify/suite"
)

type prometheusTestSuite struct {
	suite.Suite
}

func (s *prometheusTestSuite) TestNew() {
	s.Assert().NotNil(prometheus.New())
}

func TestPrometheus(t *testing.T) {
	suite.Run(t, new(prometheusTestSuite))
}
