package metrics_test

import (
	"testing"
	"time"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/libraries/metrics"
	"wikimedia-enterprise/api/main/submodules/httputil"

	"github.com/stretchr/testify/suite"
)

type metricsTestSuite struct {
	suite.Suite
	env *env.Environment
}

func (s *metricsTestSuite) SetupSuite() {
	s.env = new(env.Environment)
	s.env.PrometheusPort = 100
	s.env.MetricsReadTimeOutSeconds = 10
	s.env.MetricsReadHeaderTimeOutSeconds = 10
	s.env.MetricsWriteTimeoutSeconds = 10
}

func (s *metricsTestSuite) TestNew() {
	mrr := metrics.New(s.env).(*httputil.MetricsRecorder)

	s.Assert().NotNil(mrr)
	s.Assert().Equal(s.env.PrometheusPort, mrr.ServerPort)
	s.Assert().Equal(time.Duration(s.env.MetricsReadHeaderTimeOutSeconds)*time.Second, mrr.Server.ReadHeaderTimeout)
	s.Assert().Equal(time.Duration(s.env.MetricsReadTimeOutSeconds)*time.Second, mrr.Server.ReadTimeout)
	s.Assert().Equal(time.Duration(s.env.MetricsWriteTimeoutSeconds)*time.Second, mrr.Server.WriteTimeout)
}

func TestMetrics(t *testing.T) {
	suite.Run(t, new(metricsTestSuite))
}
