package metrics

import "github.com/prometheus/client_golang/prometheus"

func registerCounterVec(reg prometheus.Registerer, name string, labels []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: "Count of " + name,
		},
		labels,
	)
	reg.MustRegister(c)
	return c
}

func registerDurationHistogram(reg prometheus.Registerer, name string, labels []string) *prometheus.HistogramVec {
	h := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    "Duration histogram of " + name,
			Buckets: prometheus.DefBuckets,
		},
		labels,
	)
	reg.MustRegister(h)
	return h
}
