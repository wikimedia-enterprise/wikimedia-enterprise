// Package prometheus provides instumentation toolkit for Prometheus.
package prometheus

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redis "github.com/redis/go-redis/v9"

	redisprometheus "github.com/redis/go-redis/extra/redisprometheus/v9"
)

// Metrics endpoint default URL.
const metricsURL = "/metrics"

// Parameters required for this toolkit.
type Parameters struct {
	Port       int
	Collectors []prometheus.Collector
	Metrics    *Metrics
	Redis      redis.Cmdable
	Srv        *http.Server
}

// addDefaultCollectors function appends default collectors.
func addDefaultCollectors(cs []prometheus.Collector) []prometheus.Collector {
	// Add default collectors.
	return append(cs,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// addRedisCollector function appends default Redis collector.
func addDefaultRedisCollector(cs []prometheus.Collector, r *redis.Client) []prometheus.Collector {
	// Add default collectors.
	return append(cs, redisprometheus.NewCollector("", "", r))
}

// NewServer creats a new server.
func NewServer(p int, handler http.Handler) *http.Server {
	http.DefaultServeMux.Handle(metricsURL, handler)
	// handler is nil to invoke http.DefaultServeMux
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%v", p),
		Handler:           nil,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
	}
	return srv
}

// Runs a server with the endpoint /metrics for Prometheus.
func Run(p Parameters) error {
	p.Collectors = addDefaultCollectors(p.Collectors)

	if p.Redis != nil {
		p.Collectors = addDefaultRedisCollector(p.Collectors, p.Redis.(*redis.Client))
	}

	// Create non-global registry.
	reg := prometheus.NewRegistry()

	// Registering collectors.
	reg.MustRegister(p.Collectors...)

	if err := p.Metrics.RegisterMetrics(reg); err != nil {
		log.Println(err)
		return err
	}

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg, EnableOpenMetrics: true})

	if p.Srv == nil {
		p.Srv = NewServer(p.Port, handler)
	}

	if err := p.Srv.ListenAndServe(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
