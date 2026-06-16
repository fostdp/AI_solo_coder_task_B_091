package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		if path != "" {
			ObserveHTTPRequest(c.Request.Method, path, status)
			ObserveHTTPDuration(c.Request.Method, path, duration)
		}
	}
}
