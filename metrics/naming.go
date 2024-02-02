package metrics

const appName = "pulse"
const requestsMetricBaseName = "requests"

func metricName(name string) string {
	return "bux_" + name
}

func counterName(name string) string {
	return metricName(name) + "_total"
}

func durationSecName(name string) string {
	return metricName(name) + "_duration_seconds"
}
