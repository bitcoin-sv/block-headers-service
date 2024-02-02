package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type RequestMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func registerRequestMetrics(reg *prometheus.Registry) *RequestMetrics {
	requestsTotal := registerCounterVec(reg, requestsMetricBaseName, []string{"method", "path", "status", "classification"})
	requestDuration := registerDurationHistogram(reg, requestsMetricBaseName, []string{"method", "path"})

	return &RequestMetrics{
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
	}
}

func (m *RequestMetrics) Track(method, path string) *RequestTracker {
	return &RequestTracker{
		method:  method,
		path:    path,
		metrics: m,
	}
}

type RequestTracker struct {
	method    string
	path      string
	startTime time.Time
	metrics   *RequestMetrics
}

func (r *RequestTracker) Start() {
	r.startTime = time.Now()
}

func (r *RequestTracker) End(status int) {
	if status == 404 {
		// This is a safeguard against attacks where the server is flooded with requests having unique paths,
		// which would lead to the creation of a large number of metrics
		r.path = "UNKNOWN_PATH"
	}
	r.metrics.requestsTotal.WithLabelValues(r.method, r.path, fmt.Sprint(status), requestClassification(status)).Inc()
	r.metrics.requestDuration.WithLabelValues(r.method, r.path).Observe(time.Since(r.startTime).Seconds())
}

func requestClassification(status int) string {
	if status >= 200 && status < 400 {
		return "success"
	}
	return "failure"
}
