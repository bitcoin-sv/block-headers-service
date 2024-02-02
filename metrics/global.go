package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metrics *Metrics

func EnableMetrics() {
	metrics = newMetrics()
}

func Get() (m *Metrics, enabled bool) {
	return metrics, metrics != nil
}

func Register(ginEngine *gin.Engine) {
	if metrics, enabled := Get(); enabled {
		ginEngine.Use(requestMetricsMiddleware())

		rootGroup := ginEngine.Group("/metrics")

		httpHandler := promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{Registry: metrics.registry})
		rootGroup.GET("", gin.WrapH(httpHandler))
	}
}
