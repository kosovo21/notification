package model

import (
	"time"

	"github.com/google/uuid"
)

// MessageStatus represents the processing state of a message.
type MessageStatus int

const (
	StatusQueued     MessageStatus = 0
	StatusProcessing MessageStatus = 1
	StatusSent       MessageStatus = 2
	StatusDelivered  MessageStatus = 3
	StatusFailed     MessageStatus = 4
	StatusPending    MessageStatus = 5 // will retry
	StatusCancelled  MessageStatus = 6
)

// String returns the human-readable name of the status.
func (s MessageStatus) String() string {
	switch s {
	case StatusQueued:
		return "queued"
	case StatusProcessing:
		return "processing"
	case StatusSent:
		return "sent"
	case StatusDelivered:
		return "delivered"
	case StatusFailed:
		return "failed"
	case StatusPending:
		return "pending"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Platform represents a notification delivery channel.
type Platform string

const (
	PlatformSMS      Platform = "sms"
	PlatformWhatsApp Platform = "whatsapp"
	PlatformTelegram Platform = "telegram"
	PlatformEmail    Platform = "email"
)

// Priority represents the urgency level of a message.
type Priority int

const (
	PriorityLow    Priority = 0
	PriorityNormal Priority = 1
	PriorityHigh   Priority = 2
)

// Message represents a notification message stored in the database.
type Message struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	UserID      uuid.UUID     `json:"user_id" db:"user_id"`
	Subject     string        `json:"subject" db:"subject"`
	Body        string        `json:"body" db:"body"`
	Sender      string        `json:"sender" db:"sender"`
	Platform    Platform      `json:"platform" db:"platform"`
	Priority    Priority      `json:"priority" db:"priority"`
	Status      MessageStatus `json:"status" db:"status"`
	ScheduledAt *time.Time    `json:"scheduled_at,omitempty" db:"scheduled_at"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}
