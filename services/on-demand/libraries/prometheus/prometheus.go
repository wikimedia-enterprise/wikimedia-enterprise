package prometheus

import (
	pr "wikimedia-enterprise/general/prometheus"
)

// NewAPI creates a new prometheus API.
func NewAPI() *pr.Metrics {
	prm := new(pr.Metrics)
	prm.Init()
	return prm
}
