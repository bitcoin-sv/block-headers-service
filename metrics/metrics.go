package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var LabelName = "key"

type Metrics struct {
	registry     *prometheus.Registry
	httpRequests *RequestMetrics
}

func newMetrics() *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		registry:     reg,
		httpRequests: registerRequestMetrics(reg),
	}

	return m
}
