package metrics

import "github.com/prometheus/client_golang/prometheus"

func registerCounterVec(reg *prometheus.Registry, baseName string, labels []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: counterName(baseName),
			Help: "Count of " + baseName,
		},
		labels,
	)
	reg.MustRegister(c)
	return c
}

func registerDurationHistogram(reg *prometheus.Registry, baseName string, labels []string) *prometheus.HistogramVec {
	h := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    durationSecName(baseName),
			Help:    "Duration histogram of " + baseName,
			Buckets: prometheus.DefBuckets,
		},
		labels,
	)
	reg.MustRegister(h)
	return h
}
