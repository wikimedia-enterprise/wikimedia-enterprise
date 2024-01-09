package httputil

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsDefaultValueChannelCap sets the default capacity of the metrics value channel.
const MetricsDefaultValueChannelCap = 10000

// MetricsEventType represent available event types for metric events.
type MetricsEventType int

// These constants are used to represent the type of metrics event.
const (
	MetricsInitEvent MetricsEventType = 0
	MetricsPushEvent MetricsEventType = 1
)

// Labels to be used by the metrics.
const (
	MetricsLabelMethod    = "method"
	MetricsLabelPath      = "path"
	MetricsLabelStatus    = "status"
	MetricsLabelUsername  = "username"
	MetricsLabelUserGroup = "user_group"
	MetricsLabelClientIP  = "client_ip"
)

// MetricsDefaultLabels contains the labels to be attached to a metrics value by default.
var MetricsDefaultLabels = []string{
	MetricsLabelMethod,
	MetricsLabelPath,
	MetricsLabelStatus,
	MetricsLabelUsername,
	MetricsLabelUserGroup,
	MetricsLabelClientIP,
}

// MetricsOpenConnectionLabels contains the labels to be attached to a metrics value for open connections.
// Note that this is a subset of the MetricsDefaultLabels, cuz we are narrowing it down.
var MetricsOpenConnectionLabels = []string{
	MetricsLabelMethod,
	MetricsLabelPath,
	MetricsLabelUsername,
	MetricsLabelUserGroup,
	MetricsLabelClientIP,
}

// NewMetricsValue creates a new metrics value with start time initialized.
// The functionality can be repeated by calling SetStartTime manually.
// This is just a convenience function.
func NewMetricsValue() MetricsValue {
	mvl := MetricsValue{}
	mvl.SetStartTime()
	return mvl
}

// MetricsValue represents a metrics event.
type MetricsValue struct {
	eventType MetricsEventType
	StartTime time.Time
	Labels    []string
}

// SetLabels sets the labels for a metrics value based on the given gin context.
func (m *MetricsValue) SetLabels(gcx *gin.Context) {
	usr := NewUser(gcx)

	m.Labels = []string{
		gcx.Request.Method,
		gcx.Request.URL.Path,
		strconv.Itoa(gcx.Writer.Status()),
		usr.GetUsername(),
		usr.GetGroup(),
		gcx.ClientIP(),
	}
}

// GetLabels will return values for the labels.
// It has an optional parameter to get only the subset of values,
// and will return the default label values if no parameters are provided.
func (m *MetricsValue) GetLabels(kls ...string) []string {
	// if no specific label keys are provided, returning the default set of labels
	if len(kls) == 0 {
		return m.Labels
	}

	lbs := []string{}

	// we are creating a subset of labels here
	for _, key := range kls {
		for idx, dky := range MetricsDefaultLabels {
			if dky == key {
				lbs = append(lbs, m.Labels[idx])
				break
			}
		}
	}

	return lbs
}

// SetEventType setter function for the event type.
func (m *MetricsValue) SetEventType(evt MetricsEventType) {
	m.eventType = evt
}

// GetEventType getter for event type private property.
func (m *MetricsValue) GetEventType() MetricsEventType {
	return m.eventType
}

// SetStartTime sets the start time for a metrics value.
func (m *MetricsValue) SetStartTime() {
	m.StartTime = time.Now().UTC()
}

// MetricsIniter is an interface that specifies a method to initialize a metrics value.
type MetricsIniter interface {
	Init(MetricsValue)
}

// MetricsPusher is an interface that specifies a method to push a metrics value.
type MetricsPusher interface {
	Push(MetricsValue)
}

// MetricsServer is an interface that specifies a method to serve metrics data.
type MetricsServer interface {
	Serve() error
}

// MetricsRecorderAPI is an interface that combines MetricsIniter and MetricsPusher.
type MetricsRecorderAPI interface {
	MetricsIniter
	MetricsPusher
	MetricsServer
}

// NewMetricsRecorder creates and returns a new instance of MetricsRecorderAPI.
// You can use options function  to customize the recorder, argument is optional.
func NewMetricsRecorder(ops ...func(mrr *MetricsRecorder)) MetricsRecorderAPI {
	mrr := &MetricsRecorder{
		ServerPort: 12411,
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			MetricsDefaultLabels,
		),
		HTTPRequestsDuration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_duration_seconds",
				Help: "HTTP request duration in seconds.",
			},
			MetricsDefaultLabels,
		),
		HTTPOpenConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_open_connections",
				Help: "HTTP number of open connections.",
			},
			MetricsOpenConnectionLabels,
		),
	}

	for _, opt := range ops {
		opt(mrr)
	}

	return mrr
}

// MetricsRecorder is a struct that holds the configuration and metrics of a monitoring system.
type MetricsRecorder struct {
	Server               *http.Server
	ServerPort           int
	Values               chan MetricsValue
	ValuesChannelCap     int
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestsDuration *prometheus.SummaryVec
	HTTPOpenConnections  *prometheus.GaugeVec
}

// Init initializes a MetricsValue instance and pushes it onto the Values channel.
func (m *MetricsRecorder) Init(mvl MetricsValue) {
	mvl.SetEventType(MetricsInitEvent)
	m.Values <- mvl
}

// Push pushes a MetricsValue instance onto the Values channel.
func (m *MetricsRecorder) Push(mvl MetricsValue) {
	mvl.SetEventType(MetricsPushEvent)
	m.Values <- mvl
}

// Serve starts an HTTP server that listens on the specified port and serves metrics data.
func (m *MetricsRecorder) Serve() error {
	reg := prometheus.NewRegistry()

	for _, clt := range []prometheus.Collector{
		collectors.NewGoCollector(),
		m.HTTPRequestsTotal,
		m.HTTPRequestsDuration,
		m.HTTPOpenConnections,
	} {
		if err := reg.Register(clt); err != nil {
			return err
		}
	}

	if m.ValuesChannelCap <= 0 {
		m.ValuesChannelCap = MetricsDefaultValueChannelCap
	}

	if m.Values == nil {
		m.Values = make(chan MetricsValue, m.ValuesChannelCap)
	}

	go func() {
		for mvl := range m.Values {
			switch mvl.GetEventType() {
			case MetricsInitEvent:
				m.HTTPOpenConnections.WithLabelValues(mvl.GetLabels(MetricsOpenConnectionLabels...)...).Inc()
			case MetricsPushEvent:
				m.HTTPOpenConnections.WithLabelValues(mvl.GetLabels(MetricsOpenConnectionLabels...)...).Dec()
				m.HTTPRequestsTotal.WithLabelValues(mvl.GetLabels()...).Inc()
				m.HTTPRequestsDuration.WithLabelValues(mvl.GetLabels()...).Observe(float64(time.Since(mvl.StartTime).Seconds()))
			}
		}
	}()

	if m.Server == nil {
		m.Server = new(http.Server)
	}

	if len(m.Server.Addr) == 0 {
		m.Server.Addr = fmt.Sprintf(":%d", m.ServerPort)
	}

	if m.Server.Handler == nil {
		hdr := http.NewServeMux()

		hdr.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			Registry:          reg,
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
// It sets up a MetricsValue struct to hold information about the request, initializes the MetricsRecorderAPI with the MetricsValue,
// calls the next handler in the chain, and then pushes the MetricsValue to the MetricsRecorderAPI for recording.
func Metrics(mrr MetricsRecorderAPI) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mvl := NewMetricsValue()
		mvl.SetLabels(gcx)
		mrr.Init(mvl)

		defer func() {
			mvl.SetLabels(gcx)
			mrr.Push(mvl)
		}()

		gcx.Next()
	}
}
