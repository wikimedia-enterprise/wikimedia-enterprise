// Package metrics provides a wrapper for dependency injection of httputil.MetricsRecorder.
package metrics

import (
	"net/http"
	"time"
	"wikimedia-enterprise/api/main/config/env"
	"wikimedia-enterprise/api/main/submodules/httputil"
)

// New creates new instance of metrics recorder API.
// Sets the server prot to the prometheus port for the environment variable.
func New(env *env.Environment) httputil.MetricsRecorderAPI {
	return httputil.NewMetricsRecorder(func(mrr *httputil.MetricsRecorder) {
		mrr.ServerPort = env.PrometheusPort
		// Sets the metrics server timeouts
		mrr.Server = &http.Server{
			ReadHeaderTimeout: time.Duration(env.MetricsReadHeaderTimeOutSeconds) * time.Second,
			ReadTimeout:       time.Duration(env.MetricsReadTimeOutSeconds) * time.Second,
			WriteTimeout:      time.Duration(env.MetricsWriteTimeoutSeconds) * time.Second,
		}
	})
}
