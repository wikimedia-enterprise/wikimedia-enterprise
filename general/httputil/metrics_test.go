package httputil

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type metricsRecorderTestSuite struct {
	suite.Suite
	mdr            *MetricsRecorder
	srv            *httptest.Server
	requestCounter int32
}

func (s *metricsRecorderTestSuite) SetupTest() {
	atomic.StoreInt32(&s.requestCounter, 0)
}

func (s *metricsRecorderTestSuite) SetupSuite() {
	opt := func(mrr *MetricsRecorder) {
		s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&s.requestCounter, 1)
			if r.URL.Path == "/metrics" {
				if count > 3 {
					w.WriteHeader(http.StatusTooManyRequests)
					fmt.Fprintf(w, "Too many requests")
					return
				}
			} else {
				http.NotFound(w, r)
				fmt.Fprintf(w, "Metrics")
			}
		}))
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

func (s *metricsRecorderTestSuite) TestServe() {
	res, err := http.Get(fmt.Sprintf("%s/metrics", s.srv.URL))
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	dta, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Empty(string(dta))
}

func TestMetricsRecorder(t *testing.T) {
	suite.Run(t, new(metricsRecorderTestSuite))
}

func TestMetricsMiddleware(t *testing.T) {
	assert := assert.New(t)

	rdr := NewMetricsRecorder()
	mwr := Metrics(rdr)

	gin.SetMode(gin.TestMode)
	rtr := gin.New()
	rtr.GET("/path", mwr, func(ctx *gin.Context) {
		ctx.Writer.WriteHeader(201)
		assert.Equal(1.0, testutil.ToFloat64(HTTPOpenConnections))
	})
	srv := httptest.NewServer(rtr)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/path")
	assert.NoError(err)
	defer resp.Body.Close()

	assert.Equal(0.0, testutil.ToFloat64(HTTPOpenConnections))
	assert.Equal(1.0, testutil.ToFloat64(HTTPRequestsTotal))
	assert.Equal(1.0, testutil.ToFloat64(HTTPRequestStatusTotal.WithLabelValues("GET", "/path", "201")))
}
