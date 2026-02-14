package model

import (
	"time"

	"github.com/google/uuid"
)

// Recipient represents a single delivery target for a message.
type Recipient struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	MessageID    uuid.UUID     `json:"message_id" db:"message_id"`
	Recipient    string        `json:"recipient" db:"recipient"`
	Status       MessageStatus `json:"status" db:"status"`
	ProviderID   *string       `json:"provider_id,omitempty" db:"provider_id"`
	ErrorMessage *string       `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int           `json:"retry_count" db:"retry_count"`
	SentAt       *time.Time    `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt  *time.Time    `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}
