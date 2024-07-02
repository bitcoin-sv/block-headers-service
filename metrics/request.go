package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RequestMetrics is a collection of metrics related to HTTP requests.
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

// Track returns a RequestTracker to track the duration and status of an HTTP request.
func (m *RequestMetrics) Track(method, path string) *RequestTracker {
	return &RequestTracker{
		method:  method,
		path:    path,
		metrics: m,
	}
}

// RequestTracker is a helper struct to track the duration and status of an HTTP request.
type RequestTracker struct {
	method    string
	path      string
	startTime time.Time
	metrics   *RequestMetrics
}

// Start marks the beginning of the request.
func (r *RequestTracker) Start() {
	r.startTime = time.Now()
}

// End marks the end of the request and writes the duration and status to the metrics.
func (r *RequestTracker) End(status int) {
	r.writeCounter(status, r.path)
	r.writeDuration()
}

// EndWithNoRoute marks the end of the request with a 404 status and writes the duration to the metrics.
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
