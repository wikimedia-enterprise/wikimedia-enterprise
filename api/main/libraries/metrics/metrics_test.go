package metrics_test

import (
	"testing"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/libraries/metrics"
	"wikimedia-enterprise/general/httputil"

	"github.com/stretchr/testify/suite"
)

type metricsTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *metricsTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.PrometheusPort = 100
}

func (s *metricsTestSuite) TestNew() {
	mrr := metrics.New(s.env).(*httputil.MetricsRecorder)

	s.Assert().NotNil(mrr)
	s.Assert().Equal(s.env.PrometheusPort, mrr.ServerPort)
}

func TestMetrics(t *testing.T) {
	suite.Run(t, new(metricsTestSuite))
}
