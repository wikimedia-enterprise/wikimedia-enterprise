package httputil

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Labels to be used by the metrics.
const (
	MetricsLabelMethod = "method"
	MetricsLabelPath   = "path"
	MetricsLabelStatus = "status"
)

// MetricsDefaultLabels contains the labels to be attached to a metrics value by default.
var MetricsDefaultLabels = []string{
	MetricsLabelMethod,
	MetricsLabelPath,
	MetricsLabelStatus,
}

// MetricsDefaultLabels contains the labels available at the start of a request.
var MetricsRequestStartLabels = []string{
	MetricsLabelMethod,
	MetricsLabelPath,
}

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, MetricsRequestStartLabels)
	HTTPRequestStatusTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_status_total",
		Help: "Total number of HTTP requests with their status code",
	}, MetricsDefaultLabels)
	HTTPRequestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_latency_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: []float64{.2, .4, .6, .8, 1, 2, 5, 10, 30, 60, 120},
	}, MetricsRequestStartLabels)
	HTTPOpenConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_open_connections",
		Help: "HTTP number of open connections.",
	}, MetricsRequestStartLabels)
)

// MetricsRecorderAPI is an interface implemented by MetricsRecorder.
type MetricsRecorderAPI interface {
	MetricsServer
}

// GetLabels returns the values of the requested labels, or all if none are provided.
func getLabels(gcx *gin.Context, labels ...string) []string {
	// Order here needs to match the order in which the labels are defined in each metric.
	vls := []string{
		gcx.Request.Method,
		gcx.FullPath(),
		strconv.Itoa(gcx.Writer.Status()),
	}

	if len(labels) == 0 {
		return vls
	}

	out := []string{}
	for _, key := range labels {
		for idx, dky := range MetricsDefaultLabels {
			if dky == key {
				out = append(out, vls[idx])
				break
			}
		}
	}

	return out
}

// MetricsServer is an interface that specifies a method to serve metrics data.
type MetricsServer interface {
	Serve() error
}

// NewMetricsRecorder creates and returns a new instance of MetricsRecorderAPI.
// You can use options function to customize the recorder, argument is optional.
func NewMetricsRecorder(ops ...func(mrr *MetricsRecorder)) MetricsRecorderAPI {
	mrr := &MetricsRecorder{
		ServerPort: 12411,
	}

	for _, opt := range ops {
		opt(mrr)
	}

	return mrr
}

// MetricsRecorder is a struct that holds the configuration and metrics of a monitoring system.
type MetricsRecorder struct {
	Server     *http.Server
	ServerPort int
}

// Serve starts an HTTP server that listens on the specified port and serves metrics data.
func (m *MetricsRecorder) Serve() error {
	if m.Server == nil {
		m.Server = new(http.Server)
	}

	if len(m.Server.Addr) == 0 {
		m.Server.Addr = fmt.Sprintf(":%d", m.ServerPort)
	}

	if m.Server.Handler == nil {
		hdr := http.NewServeMux()

		hdr.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			Registry:          prometheus.DefaultRegisterer,
			EnableOpenMetrics: true,
		}))

		m.Server.Handler = hdr
	}

	if m.Server.ReadHeaderTimeout == 0 {
		m.Server.ReadHeaderTimeout = 2 * time.Second
	}

	if m.Server.ReadTimeout == 0 {
		m.Server.ReadTimeout = 2 * time.Second
	}

	if m.Server.WriteTimeout == 0 {
		m.Server.WriteTimeout = 2 * time.Second
	}

	return m.Server.ListenAndServe()
}

// Metrics returns a Gin middleware that records metrics using the provided MetricsRecorderAPI.
func Metrics(mrr MetricsRecorderAPI) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		startLabels := getLabels(gcx, MetricsRequestStartLabels...)
		HTTPRequestsTotal.WithLabelValues(startLabels...).Inc()
		timer := prometheus.NewTimer(HTTPRequestLatency.WithLabelValues(startLabels...))
		HTTPOpenConnections.WithLabelValues(startLabels...).Inc()

		defer func() {
			endLabels := getLabels(gcx)
			HTTPOpenConnections.WithLabelValues(startLabels...).Dec()
			HTTPRequestStatusTotal.WithLabelValues(endLabels...).Inc()
			timer.ObserveDuration()
		}()

		gcx.Next()
	}
}
