package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"notification-system/internal/config"
)

// TwilioAdapter sends SMS messages via the Twilio REST API.
type TwilioAdapter struct {
	accountSID  string
	authToken   string
	phoneNumber string
	httpClient  *http.Client
}

// NewTwilioAdapter creates a new TwilioAdapter.
func NewTwilioAdapter(cfg config.TwilioConfig) *TwilioAdapter {
	return &TwilioAdapter{
		accountSID:  cfg.AccountSID,
		authToken:   cfg.AuthToken,
		phoneNumber: cfg.PhoneNumber,
		httpClient:  &http.Client{},
	}
}

// Send sends an SMS via Twilio.
func (t *TwilioAdapter) Send(ctx context.Context, to, subject, body string) (*SendResult, error) {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", t.phoneNumber)
	data.Set("Body", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create twilio request: %w", err)
	}

	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twilio request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		SID          string `json:"sid"`
		ErrorCode    *int   `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode twilio response: %w", err)
	}

	if resp.StatusCode >= 400 || result.ErrorCode != nil {
		log.Error().
			Str("to", to).
			Int("status_code", resp.StatusCode).
			Str("error_message", result.ErrorMessage).
			Msg("twilio send failed")
		return nil, fmt.Errorf("twilio error: %s", result.ErrorMessage)
	}

	return &SendResult{ProviderID: result.SID}, nil
}

// Platform returns "sms".
func (t *TwilioAdapter) Platform() string {
	return "sms"
}
