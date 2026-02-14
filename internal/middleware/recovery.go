package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"notification-system/internal/model"
	"notification-system/pkg/logger"
)

// Recovery returns a middleware that recovers from panics,
// logs the stack trace, and returns a 500 JSON error.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log := logger.Get()
				log.Error().
					Interface("error", err).
					Str("stack", string(debug.Stack())).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Msg("panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
					Success: false,
					Error: model.ErrorDetail{
						Code:    "INTERNAL_ERROR",
						Message: "An unexpected error occurred",
					},
				})
			}
		}()

		c.Next()
	}
}
