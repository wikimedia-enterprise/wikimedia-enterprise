package httputil

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type metricsValueTestSuite struct {
	suite.Suite
	mvl *MetricsValue
	pth string
}

func (s *metricsValueTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.mvl = new(MetricsValue)
	s.pth = "/hello"
}

func (s *metricsValueTestSuite) TestSetLabels() {
	gcx, _ := gin.CreateTestContext(httptest.NewRecorder())
	gcx.Request = httptest.NewRequest(http.MethodGet, s.pth, nil)
	s.mvl.SetLabels(gcx)

	s.Assert().Equal(http.MethodGet, s.mvl.Labels[0])
	s.Assert().Equal(s.pth, s.mvl.Labels[1])
	s.Assert().Equal(strconv.Itoa(http.StatusOK), s.mvl.Labels[2])
	s.Assert().Empty(s.mvl.Labels[3])
	s.Assert().Empty(s.mvl.Labels[4])
	s.Assert().NotEmpty(s.mvl.Labels[5])
}

func (s *metricsValueTestSuite) TestSetStartTime() {
	s.mvl.SetStartTime()
	s.Assert().NotZero(s.mvl.StartTime)
}

func (s *metricsValueTestSuite) TestSetEventType() {
	s.mvl.SetEventType(MetricsPushEvent)
	s.Assert().Equal(MetricsPushEvent, s.mvl.GetEventType())
}

func (s *metricsValueTestSuite) TestGetLabels() {
	s.mvl.Labels = []string{"GET", "/v2/test", "200", "test", "group_3", "127.0.0.1"}
	lbs := s.mvl.GetLabels()

	s.Assert().Equal(len(s.mvl.Labels), len(lbs))
	s.Assert().Equal(s.mvl.Labels, s.mvl.GetLabels())
	s.Assert().Equal([]string{"GET", "/v2/test"}, s.mvl.GetLabels(MetricsLabelMethod, MetricsLabelPath))
}

func TestMetricValue(t *testing.T) {
	suite.Run(t, new(metricsValueTestSuite))
}

type newMetricsRecorderTestSuite struct {
	suite.Suite
	mdr *MetricsRecorder
}

func (s *newMetricsRecorderTestSuite) TestNewMetricsRecorder() {
	s.mdr = NewMetricsRecorder().(*MetricsRecorder)

	s.Assert().NotNil(s.mdr)
	s.Assert().NotNil(s.mdr.HTTPRequestsTotal)
	s.Assert().NotNil(s.mdr.HTTPRequestsDuration)
	s.Assert().NotNil(s.mdr.HTTPOpenConnections)
}

func TestNewMetricsRecorder(t *testing.T) {
	suite.Run(t, new(newMetricsRecorderTestSuite))
}

type metricsRecorderTestSuite struct {
	suite.Suite
	mdr *MetricsRecorder
	srv *httptest.Server
}

func (s *metricsRecorderTestSuite) SetupTest() {
	s.mdr.Values = make(chan MetricsValue, 1)
}

func (s *metricsRecorderTestSuite) SetupSuite() {
	opt := func(mrr *MetricsRecorder) {
		s.srv = httptest.NewServer(nil)
		mrr.Server = s.srv.Config
	}

	s.mdr = NewMetricsRecorder(opt).(*MetricsRecorder)

	go func() {
		_ = s.mdr.Serve()
	}()
}

func (s *metricsRecorderTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *metricsRecorderTestSuite) TestInit() {
	mvl := MetricsValue{}
	s.mdr.Init(mvl)
	mvl.SetEventType(MetricsInitEvent)

	s.Assert().Equal(<-s.mdr.Values, mvl)
}

func (s *metricsRecorderTestSuite) TestPush() {
	mvl := MetricsValue{}
	s.mdr.Push(mvl)
	mvl.SetEventType(MetricsPushEvent)

	s.Assert().Equal(<-s.mdr.Values, mvl)
}

func (s *metricsRecorderTestSuite) TestServe() {
	res, err := http.Get(fmt.Sprintf("%s/metrics", s.srv.URL))
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().NotEmpty(string(dta))
}

func TestMetricsRecorder(t *testing.T) {
	suite.Run(t, new(metricsRecorderTestSuite))
}

type metricsRecorderMock struct {
	mock.Mock
	MetricsRecorderAPI
}

func (m *metricsRecorderMock) Init(_ MetricsValue) {
	m.Called()
}

func (m *metricsRecorderMock) Push(_ MetricsValue) {
	m.Called()
}

type metricsMiddlewareTestSuite struct {
	suite.Suite
	pth string
	mrm *metricsRecorderMock
}

func (s *metricsMiddlewareTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.pth = "/hello"
	s.mrm = new(metricsRecorderMock)
	s.mrm.On("Init").Return()
	s.mrm.On("Push").Return()
}

func (s *metricsMiddlewareTestSuite) TestMetrics() {
	gcx, _ := gin.CreateTestContext(httptest.NewRecorder())
	gcx.Request = httptest.NewRequest(http.MethodGet, s.pth, nil)

	mwr := Metrics(s.mrm)
	mwr(gcx)

	s.mrm.AssertCalled(s.T(), "Init")
	s.mrm.AssertCalled(s.T(), "Push")
}

func TestMetricsMiddleware(t *testing.T) {
	suite.Run(t, new(metricsMiddlewareTestSuite))
}
