// Package prometheus provides instumentation toolkit for Prometheus.
package prometheus

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
	if p.Redis != nil {
		p.Collectors = addDefaultRedisCollector(p.Collectors, p.Redis.(*redis.Client))
	}

	// Registering collectors.
	prometheus.MustRegister(p.Collectors...)

	if err := p.Metrics.RegisterMetrics(prometheus.DefaultRegisterer.(*prometheus.Registry)); err != nil {
		log.Println(err)
		return err
	}

	handler := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{Registry: prometheus.DefaultRegisterer, EnableOpenMetrics: true})

	if p.Srv == nil {
		p.Srv = NewServer(p.Port, handler)
	}

	if err := p.Srv.ListenAndServe(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
