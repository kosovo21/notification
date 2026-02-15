package adapter

import "context"

// SendResult holds the result of a send operation.
type SendResult struct {
	ProviderID string // Provider-assigned message ID
}

// Sender defines the interface for sending notifications through a platform.
type Sender interface {
	// Send delivers a notification to the given recipient.
	Send(ctx context.Context, to, subject, body string) (*SendResult, error)

	// Platform returns the platform name this sender handles.
	Platform() string
}
