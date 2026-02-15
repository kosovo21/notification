package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// MockAdapter simulates sending notifications for local development and testing.
type MockAdapter struct {
	platform string
	delay    time.Duration
}

// NewMockAdapter creates a new MockAdapter for the given platform.
func NewMockAdapter(platform string) *MockAdapter {
	return &MockAdapter{
		platform: platform,
		delay:    100 * time.Millisecond,
	}
}

// Send simulates sending a notification by logging and returning a fake provider ID.
func (m *MockAdapter) Send(ctx context.Context, to, subject, body string) (*SendResult, error) {
	// Simulate network delay
	select {
	case <-time.After(m.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	providerID := fmt.Sprintf("mock-%s-%s", m.platform, uuid.New().String()[:8])

	log.Info().
		Str("platform", m.platform).
		Str("to", to).
		Str("subject", subject).
		Str("provider_id", providerID).
		Msg("[MOCK] notification sent")

	return &SendResult{ProviderID: providerID}, nil
}

// Platform returns the platform name.
func (m *MockAdapter) Platform() string {
	return m.platform
}
