package prometheus

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type metricsTestSuite struct {
	suite.Suite
	mts *Metrics
}

func (m *metricsTestSuite) SetupSuite() {
	m.mts = new(Metrics)
}
func (m *metricsTestSuite) TestAddMetrics() {
	m.mts.Init()
	m.Assert().NotNil(m.mts.Opts)

	m.mts.AddEventStreamMetrics()
	m.Assert().NotEmpty(m.mts.Opts[EsTtlErrs])
	m.Assert().NotEmpty(m.mts.Opts[EsTtlEvents])
	m.Assert().NotEmpty(m.mts.Opts[EsTtlEvntsPs])

	m.mts.AddRedisMetrics()
	m.Assert().NotEmpty(m.mts.Opts[RedisReqDur])
	m.Assert().NotEmpty(m.mts.Opts[RedisReqTtl])

	m.mts.AddHttpMetrics()
	m.Assert().NotEmpty(m.mts.Opts[HttpRespTime])
	m.Assert().NotEmpty(m.mts.Opts[HttpRqdr])
	m.Assert().NotEmpty(m.mts.Opts[HttpTlrq])

	m.mts.AddPerformanceMetrics()
	m.Assert().NotEmpty(m.mts.Opts[Duration])
	m.Assert().NotEmpty(m.mts.Opts[TtlErrs])

	m.mts.AddStructuredDataMetrics()
	m.Assert().NotEmpty(m.mts.Opts[SDTtlErrs])
	m.Assert().NotEmpty(m.mts.Opts[SDTtlEvnts])
	m.Assert().NotEmpty(m.mts.Opts[SDTtlEvntsPs])

	m.mts.AddStructuredContentsMetrics()
	m.Assert().NotEmpty(m.mts.Opts[SCTtlErrs])
	m.Assert().NotEmpty(m.mts.Opts[SCTtlEvents])
	m.Assert().NotEmpty(m.mts.Opts[SCTtlEvntsPs])

	m.mts.AddCommonsMetrics()
	m.Assert().NotEmpty(m.mts.Opts[CommonsTtlErrs])
	m.Assert().NotEmpty(m.mts.Opts[CommonsTtlEvents])
	m.Assert().NotEmpty(m.mts.Opts[CommonsTtlEvntsPs])

	m.mts.AddOnDemandMetrics()
	m.Assert().NotEmpty(m.mts.Opts[OdmTtlErrs])
	m.Assert().NotEmpty(m.mts.Opts[OdmTtlEvents])
}

func TestMetrics(t *testing.T) {
	suite.Run(t, new(metricsTestSuite))
}
