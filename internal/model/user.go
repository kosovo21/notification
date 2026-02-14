package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered API user.
type User struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	APIKeyHash    string    `json:"-" db:"api_key_hash"`
	Role          string    `json:"role" db:"role"`
	RateLimitTier string    `json:"rate_limit_tier" db:"rate_limit_tier"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
