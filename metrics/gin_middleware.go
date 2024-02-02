package metrics

import (
	"github.com/gin-gonic/gin"
)

func requestMetricsMiddleware() gin.HandlerFunc {
	if metrics, enabled := Get(); enabled {
		return func(c *gin.Context) {
			tracker := metrics.httpRequests.Track(c.Request.Method, c.Request.URL.Path)
			tracker.Start()
			defer tracker.End(c.Writer.Status()) //note that the status code will be correct only if another middleware doesn't change the status code; order of middlewares matters

			c.Next()
		}
	}

	return func(c *gin.Context) {
		c.Next()
	}
}
