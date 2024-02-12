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

func registerRequestMetrics(reg prometheus.Registerer) *RequestMetrics {
	requestsTotal := registerCounterVec(reg, requestCounterName, []string{"method", "path", "status", "classification"})
	requestDuration := registerDurationHistogram(reg, requestDurationSecName, []string{"method", "path"})

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
	r.writeCounter(status, r.path)
	r.writeDuration()
}

func (r *RequestTracker) EndWithNoRoute() {
	// This is a safeguard against attacks where the server is flooded with requests having unique paths,
	// which would lead to the creation of a large number of metrics
	r.writeCounter(404, "UNKNOWN_ROUTE")
}

func (r *RequestTracker) writeCounter(status int, path string) {
	r.metrics.requestsTotal.WithLabelValues(r.method, path, fmt.Sprint(status), requestClassification(status)).Inc()
}

func (r *RequestTracker) writeDuration() {
	r.metrics.requestDuration.WithLabelValues(r.method, r.path).Observe(time.Since(r.startTime).Seconds())
}

func requestClassification(status int) string {
	if status >= 200 && status < 400 {
		return "success"
	}
	return "failure"
}
