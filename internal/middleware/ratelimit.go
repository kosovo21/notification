package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"notification-system/internal/config"
	"notification-system/internal/model"
	"notification-system/pkg/logger"
)

const defaultRateLimit = 60 // fallback if tier not found in config

// RateLimitMiddleware enforces per-user, tier-based rate limiting using Redis.
func RateLimitMiddleware(rdb *redis.Client, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		user := GetUserFromContext(c)
		if user == nil {
			c.Next()
			return
		}

		// Determine rate limit for this user's tier
		limit := defaultRateLimit
		if tier, ok := cfg.Tiers[user.RateLimitTier]; ok {
			limit = tier.RequestsPerMin
		}

		// Fixed-window counter: key = ratelimit:{user_id}:{current_minute}
		now := time.Now()
		window := now.Truncate(time.Minute)
		key := fmt.Sprintf("ratelimit:%s:%d", user.ID.String(), window.Unix())

		ctx := c.Request.Context()

		// Increment counter
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			logger.Get().Error().Err(err).Msg("rate limit redis error")
			// Fail open: allow request if Redis is down
			c.Next()
			return
		}

		// Set expiry on first request in window
		if count == 1 {
			rdb.Expire(ctx, key, 2*time.Minute)
		}

		// Calculate reset time (end of current minute window)
		resetAt := window.Add(time.Minute)
		remaining := limit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if int(count) > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, model.ErrorResponse{
				Success: false,
				Error: model.ErrorDetail{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: fmt.Sprintf("Rate limit exceeded. Limit: %d requests per minute", limit),
				},
			})
			return
		}

		c.Next()
	}
}
