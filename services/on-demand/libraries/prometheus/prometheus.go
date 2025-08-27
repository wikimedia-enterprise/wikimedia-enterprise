package prometheus

import (
	pr "wikimedia-enterprise/services/on-demand/submodules/prometheus"
)

// NewAPI creates a new prometheus API.
func NewAPI() *pr.Metrics {
	prm := new(pr.Metrics)
	prm.Init()
	return prm
}
