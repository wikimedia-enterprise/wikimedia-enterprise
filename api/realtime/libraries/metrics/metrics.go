// Package metrics provides a wrapper for dependency injection of httputil.MetricsRecorder.
package metrics

import (
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/general/httputil"
)

// New creates new instance of metrics recorder API.
// Sets the server prot to the prometheus port for the environment variable.
func New(env *env.Environment) httputil.MetricsRecorderAPI {
	return httputil.NewMetricsRecorder(func(mrr *httputil.MetricsRecorder) {
		mrr.ServerPort = env.PrometheusPort
	})
}
