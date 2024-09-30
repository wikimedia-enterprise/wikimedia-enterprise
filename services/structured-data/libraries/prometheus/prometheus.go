package prometheus

import (
	pr "wikimedia-enterprise/general/prometheus"
)

// New creates a new prometheus metrics helper instance.
func New() *pr.Metrics {
	prm := new(pr.Metrics)
	prm.Init()

	return prm
}
