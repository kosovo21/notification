package queue

import "time"

// Exchange and Routing Keys
const (
	ExchangeName       = "notification.exchange"
	RoutingKeySMS      = "notification.sms"
	RoutingKeyEmail    = "notification.email"
	RoutingKeyWhatsApp = "notification.whatsapp"
	RoutingKeyTelegram = "notification.telegram"
	RoutingKeyDLQ      = "notification.dlq"
)

// MessageQueuedEvent represents a message to be processed.
type MessageQueuedEvent struct {
	MessageID   string            `json:"message_id"`
	RecipientID string            `json:"recipient_id"`
	To          string            `json:"to"`
	Body        string            `json:"body"`
	Subject     string            `json:"subject,omitempty"`
	Platform    string            `json:"platform"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}
