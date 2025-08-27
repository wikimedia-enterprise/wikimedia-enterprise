// Package metrics provides a wrapper for dependency injection of httputil.MetricsRecorder.
package metrics

import (
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/submodules/httputil"
)

// New creates new instance of metrics recorder API.
// Sets the server port to the prometheus port for the environment variable.
func New(env *env.Environment) httputil.MetricsRecorderAPI {
	return httputil.NewMetricsRecorder(func(mrr *httputil.MetricsRecorder) {
		mrr.ServerPort = env.PrometheusPort
	})
}
