package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"notification-system/pkg/logger"
)

// Logger returns a middleware that logs each HTTP request using zerolog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		log := logger.Get()

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Int("status", status).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Dur("latency", latency).
			Int("body_size", c.Writer.Size()).
			Msg("request")
	}
}
