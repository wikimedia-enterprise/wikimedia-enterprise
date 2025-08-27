// Package metrics provides a wrapper for dependency injection of httputil.MetricsRecorder.
package metrics

import (
	"context"
	"time"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/submodules/httputil"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics for this service
var (
	Labels = []string{"has_fields", "has_filters", "has_since", "has_offsets", "has_since_per_partition", "schema", "handler", "username"}

	MessagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_delivered_total",
		Help: "The total number of messages successfully delivered",
	}, Labels)
	FirstMessageLatencySeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "first_message_delay_s",
		Help:    "The delay in seconds between establishing a connection to realtime API and delivering the first message",
		Buckets: []float64{.2, .4, .6, .8, 1, 2, 5, 10, 30, 60, 120},
	}, Labels)
	FilterLatencySeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "filter_latency_s",
		Help:    "filtering delay in seconds",
		Buckets: []float64{.2, .4, .6, .8, 1, 2, 5, 10, 30, 60, 120},
	}, Labels)
	ModifyLatencySeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "modify_latency_s",
		Help:    "message modify delay in seconds",
		Buckets: []float64{.2, .4, .6, .8, 1, 2, 5, 10, 30, 60, 120},
	}, Labels)
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total number of requests received",
	}, Labels)
	OpenConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "open_connections",
		Help: "Number of open connections",
	}, Labels)
	ConnectionTimeIdleSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "connection_time_idle_seconds",
		Help:    "The time in seconds the connection has been open without receiving new messages",
		Buckets: []float64{.2, .4, .6, .8, 1, 2, 5, 10, 30, 60, 120},
	}, Labels)
	PerConnectionThroughputSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "per_connection_throughput",
		Help:    "The number of messages per second delivered recently for a given connection",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 50, 100},
	}, Labels)
)

// New creates new instance of metrics recorder API.
// Sets the server port to the prometheus port for the environment variable.
func New(env *env.Environment) httputil.MetricsRecorderAPI {
	return httputil.NewMetricsRecorder(func(mrr *httputil.MetricsRecorder) {
		mrr.ServerPort = env.PrometheusPort
	})
}

// throughputTracker keeps track of throughput in a rolling time window. It's used to compute per-connection throughput and record
// it in Prometheus.
type throughputTracker struct {
	metric      *prometheus.HistogramVec
	labelValues []string
	samples     []int
	index       int
	count       int
}

// RecordMessage must be called for every message being tracked.
func (t *throughputTracker) RecordMessage() {
	t.count++
}

// NewThroughputTracker constructs a throughputTracker instance.
func NewThroughputTracker(windowSize int, metric *prometheus.HistogramVec, labelValues []string) *throughputTracker {
	return &throughputTracker{
		metric:      metric,
		labelValues: labelValues,
		samples:     make([]int, windowSize),
	}
}

// StartTracking sets off the rolling window tracking in a new goroutine.
func (t *throughputTracker) StartTracking(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.samples[t.index] = t.count
				t.count = 0
				t.index = (t.index + 1) % len(t.samples)

				sum := 0
				for _, v := range t.samples {
					sum += v
				}
				avg := float64(sum) / float64(len(t.samples))

				t.metric.WithLabelValues(t.labelValues...).Observe(avg)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// idleTracker periodically records the "time since last message" in the provided metric.
type idleTracker struct {
	metric      *prometheus.HistogramVec
	labelValues []string
	lastMsgTime time.Time
}

// RecordMessage must be called for every message being tracked.
func (t *idleTracker) RecordMessage() {
	t.lastMsgTime = time.Now()
}

// NewIdleTracker constructs a idleTracker instance.
func NewIdleTracker(metric *prometheus.HistogramVec, labelValues []string) *idleTracker {
	return &idleTracker{
		metric:      metric,
		labelValues: labelValues,
		lastMsgTime: time.Now(),
	}
}

// StartTracking sets off the periodic monitoring in a new goroutine.
func (t *idleTracker) StartTracking(ctx context.Context) {
	go func() {
		freq := time.Second * 15
		tmr := time.NewTicker(freq)
		defer tmr.Stop()
		for {
			select {
			case <-tmr.C:
				t.metric.WithLabelValues(t.labelValues...).Observe(float64(time.Since(t.lastMsgTime).Seconds()))
				tmr.Reset(freq)
			case <-ctx.Done():
				return
			}
		}
	}()
}
