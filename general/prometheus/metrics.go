package prometheus

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// Defined custom metrics.
const (
	// Event stream metrics.
	// Total errors.
	EsTtlErrs string = "es_ttl_errs"
	// Total events.
	EsTtlEvents string = "es_ttl_events"
	// Average events per second.
	EsTtlEvntsPs string = "es_ttl_evnts_ps"

	// Http related metrics.
	// Total requests.
	HttpTlrq string = "http_total_request"
	// Request duration.
	HttpRqdr string = "http_req_duration"
	// Responce time.
	HttpRespTime string = "http_responce_time"

	// Custom performance metrics.
	// Execution duration.
	Duration string = "duration"
	// Total errors.
	TtlErrs string = "ttl_errs"

	// Redis metrics.
	RedisReqTtl string = "redis_ttl_rqsts"
	RedisReqDur string = "redis_rqst_dur"
)

// Metrics houses all custom metrics.
type Metrics struct {
	Opts map[string]any
}

// Inc increments counter or gauge metric and vector.
func (m *Metrics) Inc(nm string, labels ...string) {
	if v, ok := m.Opts[nm].(prometheus.Counter); ok {
		v.Inc()
	}

	if v, ok := m.Opts[nm].(prometheus.Gauge); ok {
		v.Inc()
	}

	if v, ok := m.Opts[nm].(prometheus.CounterVec); ok {
		v.WithLabelValues(labels...).Inc()
	}

	if v, ok := m.Opts[nm].(prometheus.GaugeVec); ok {
		v.WithLabelValues(labels...).Inc()
	}
}

// Dec decreases gauge metrics or vector.
func (m *Metrics) Dec(nm string, labels ...string) {
	if v, ok := m.Opts[nm].(prometheus.Gauge); ok {
		v.Dec()
	}

	if v, ok := m.Opts[nm].(prometheus.GaugeVec); ok {
		v.WithLabelValues(labels...).Dec()
	}
}

// Add adds f value to gauge metric or vector.
func (m *Metrics) Add(nm string, f float64, labels ...string) {
	if v, ok := m.Opts[nm].(prometheus.Gauge); ok {
		v.Add(f)
	}

	if v, ok := m.Opts[nm].(prometheus.GaugeVec); ok {
		v.WithLabelValues(labels...).Add(f)
	}
}

// Set sets value for gauge metric or vector.
func (m *Metrics) Set(nm string, f float64, labels ...string) {
	if v, ok := m.Opts[nm].(prometheus.Gauge); ok {
		v.Set(f)
	}

	if v, ok := m.Opts[nm].(prometheus.GaugeVec); ok {
		v.WithLabelValues(labels...).Set(f)
	}
}

// Observe method add obesrvation to histrogram or summary metric or vector.
func (m *Metrics) Observe(nm string, d float64, labels ...string) {
	if v, ok := m.Opts[nm].(prometheus.Histogram); ok {
		v.Observe(d)
	}

	if v, ok := m.Opts[nm].(prometheus.Summary); ok {
		v.Observe(d)
	}

	if v, ok := m.Opts[nm].(prometheus.HistogramVec); ok {
		v.WithLabelValues(labels...).Observe(d)
	}

	if v, ok := m.Opts[nm].(prometheus.SummaryVec); ok {
		v.WithLabelValues(labels...).Observe(d)
	}
}

// Init initilizes metrics map.
func (m *Metrics) Init() {
	m.Opts = make(map[string]any)
}

// AddRedisMetrics adds RED-method Redis metrics.
func (m *Metrics) AddRedisMetrics() {
	m.Opts[RedisReqTtl] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_requests_total",
			Help: "How many Redis requests processed, partitioned by status",
		},
		[]string{"status"},
	)

	m.Opts[RedisReqDur] = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "redis_request_durations",
			Help:       "Redis requests latencies in microseconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})
}

// AddEventStreamMetrics adds event stream metrics.
func (m *Metrics) AddEventStreamMetrics() {
	m.Opts[EsTtlErrs] = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "event_stream_total_errors",
		Help: "Total number of errors",
	})

	m.Opts[EsTtlEvents] = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "event_stream_total_events",
		Help: "Total number of events",
	})

	m.Opts[EsTtlEvntsPs] = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "event_stream_average_events_per_second",
		Help: "Average number of events per second.",
	})
}

// AddPerformanceMetrics adds custom perfromance and execution-related metrics.
func (m *Metrics) AddPerformanceMetrics() {
	m.Opts[TtlErrs] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors.",
		},
		[]string{"type"},
	)

	m.Opts[Duration] = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "duration_seconds",
			Help:    "Code execution duration",
			Buckets: prometheus.DefBuckets,
		})
}

// AddHttpMetrics adds additional RED-method metrics related to HTTP.
func (m *Metrics) AddHttpMetrics() {
	m.Opts[HttpTlrq] = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of requests.",
	})

	m.Opts[HttpRqdr] = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "A histogram of the HTTP request durations in seconds.",
		Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
	})

	m.Opts[HttpRespTime] = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "namespace",
		Name:      "http_response_time",
		Help:      "Histogram of response time for a handler in seconds",
		Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"route", "method", "status_code"})

	//TODO
	//http_request_size_bytes
	//http_response_size_bytes
}

// RegisterEventStreamMetrics registers metrics.
func (m *Metrics) RegisterMetrics(reg *prometheus.Registry) error {
	for i := range m.Opts {
		if err := reg.Register(m.Opts[i].(prometheus.Collector)); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
