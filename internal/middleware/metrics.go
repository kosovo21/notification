package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"notification-system/internal/metrics"
)

// Metrics returns a Gin middleware that records Prometheus HTTP metrics.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // template path, e.g. "/api/v1/messages/:id"
		if path == "" {
			path = c.Request.URL.Path // fallback for unmatched routes
		}
		method := c.Request.Method

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
