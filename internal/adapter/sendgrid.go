package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"notification-system/internal/config"
)

// SendGridAdapter sends emails via the SendGrid v3 API.
type SendGridAdapter struct {
	apiKey     string
	fromEmail  string
	httpClient *http.Client
}

// NewSendGridAdapter creates a new SendGridAdapter.
func NewSendGridAdapter(cfg config.SendGridConfig) *SendGridAdapter {
	return &SendGridAdapter{
		apiKey:     cfg.APIKey,
		fromEmail:  cfg.FromEmail,
		httpClient: &http.Client{},
	}
}

// sendGridRequest represents the SendGrid v3 mail/send payload.
type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
}

type sendGridPersonalization struct {
	To []sendGridEmail `json:"to"`
}

type sendGridEmail struct {
	Email string `json:"email"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Send sends an email via SendGrid.
func (s *SendGridAdapter) Send(ctx context.Context, to, subject, body string) (*SendResult, error) {
	payload := sendGridRequest{
		Personalizations: []sendGridPersonalization{
			{To: []sendGridEmail{{Email: to}}},
		},
		From:    sendGridEmail{Email: s.fromEmail},
		Subject: subject,
		Content: []sendGridContent{
			{Type: "text/plain", Value: body},
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create sendgrid request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sendgrid request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Error().
			Str("to", to).
			Int("status_code", resp.StatusCode).
			Msg("sendgrid send failed")
		return nil, fmt.Errorf("sendgrid error: status %d", resp.StatusCode)
	}

	// SendGrid returns the message ID in the X-Message-Id header
	msgID := resp.Header.Get("X-Message-Id")
	if msgID == "" {
		msgID = fmt.Sprintf("sg-%s", to) // fallback
	}

	return &SendResult{ProviderID: msgID}, nil
}

// Platform returns "email".
func (s *SendGridAdapter) Platform() string {
	return "email"
}
