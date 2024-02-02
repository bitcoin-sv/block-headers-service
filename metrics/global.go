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

		ginEngine.NoRoute(func(c *gin.Context) {
			//this is needed to distinguish no-route 404 from other 404s
			c.Set(notFoundContextKey, true)
		})

		metricsGroup := ginEngine.Group("/metrics")

		httpHandler := promhttp.HandlerFor(metrics.gatherer, promhttp.HandlerOpts{Registry: metrics.registerer})
		metricsGroup.GET("", gin.WrapH(httpHandler))
	}
}
