package metrics

const serviceName = "pulse"
const requestMetricBaseName = "requests"

func metricName(name string) string {
	return serviceName + "_" + name
}

func counterName(name string) string {
	return metricName(name) + "_total"
}

func durationSecName(name string) string {
	return metricName(name) + "_duration_seconds"
}
