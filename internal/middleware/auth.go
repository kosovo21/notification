package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"notification-system/internal/auth"
	"notification-system/internal/model"
	"notification-system/internal/repository"
)

// ContextKeyUser is the Gin context key for the authenticated user.
const ContextKeyUser = "user"

// AuthMiddleware validates the X-API-Key header and sets the user in context.
func AuthMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Success: false,
				Error: model.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Missing X-API-Key header",
				},
			})
			return
		}

		hash := auth.HashAPIKey(apiKey)
		user, err := userRepo.GetByAPIKeyHash(c.Request.Context(), hash)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
					Success: false,
					Error: model.ErrorDetail{
						Code:    "UNAUTHORIZED",
						Message: "Invalid API key",
					},
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
				Success: false,
				Error: model.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to authenticate",
				},
			})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse{
				Success: false,
				Error: model.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Account is disabled",
				},
			})
			return
		}

		c.Set(ContextKeyUser, user)
		c.Next()
	}
}

// GetUserFromContext extracts the authenticated user from the Gin context.
func GetUserFromContext(c *gin.Context) *model.User {
	val, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil
	}
	user, ok := val.(*model.User)
	if !ok {
		return nil
	}
	return user
}
