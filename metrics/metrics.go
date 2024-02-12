package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	gatherer     prometheus.Gatherer
	registerer   prometheus.Registerer
	httpRequests *RequestMetrics
	latestBlock  *latestBlockMetrics
}

func newMetrics() *Metrics {
	registry := prometheus.NewRegistry()
	constLabels := prometheus.Labels{"app": appName}
	registererWithLabels := prometheus.WrapRegistererWith(constLabels, registry)

	m := &Metrics{
		gatherer:     registry,
		registerer:   registererWithLabels,
		httpRequests: registerRequestMetrics(registererWithLabels),
		latestBlock:  registerLatestBlockMetrics(registererWithLabels),
	}

	return m
}
