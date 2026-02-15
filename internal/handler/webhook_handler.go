package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"notification-system/internal/model"
	"notification-system/internal/repository"
	"notification-system/pkg/logger"
)

// WebhookHandler handles incoming provider status callbacks.
type WebhookHandler struct {
	recipientRepo repository.RecipientRepository
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(recipientRepo repository.RecipientRepository) *WebhookHandler {
	return &WebhookHandler{recipientRepo: recipientRepo}
}

// ----- Twilio -----

// twilioStatusMap maps Twilio status strings to internal MessageStatus values.
var twilioStatusMap = map[string]model.MessageStatus{
	"sent":        model.StatusSent,
	"delivered":   model.StatusDelivered,
	"undelivered": model.StatusFailed,
	"failed":      model.StatusFailed,
}

// TwilioWebhook handles POST /webhooks/twilio
// Twilio sends form-encoded callbacks with MessageSid and MessageStatus.
func (h *WebhookHandler) TwilioWebhook(c *gin.Context) {
	sid := c.PostForm("MessageSid")
	status := c.PostForm("MessageStatus")

	if sid == "" || status == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	internalStatus, ok := twilioStatusMap[strings.ToLower(status)]
	if !ok {
		// Unknown status — acknowledge but do nothing
		logger.Get().Warn().Str("status", status).Str("sid", sid).Msg("unknown twilio status, ignoring")
		c.Status(http.StatusOK)
		return
	}

	recipient, err := h.recipientRepo.GetByProviderID(c.Request.Context(), sid)
	if err != nil {
		if err == repository.ErrNotFound {
			logger.Get().Warn().Str("sid", sid).Msg("twilio webhook: recipient not found for SID")
			c.Status(http.StatusOK) // ack to prevent retries
			return
		}
		logger.Get().Error().Err(err).Str("sid", sid).Msg("twilio webhook: failed to look up recipient")
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := h.recipientRepo.UpdateStatus(c.Request.Context(), recipient.ID, internalStatus, &sid); err != nil {
		logger.Get().Error().Err(err).Str("sid", sid).Msg("twilio webhook: failed to update recipient status")
		c.Status(http.StatusInternalServerError)
		return
	}

	logger.Get().Info().
		Str("sid", sid).
		Str("status", status).
		Str("recipient_id", recipient.ID.String()).
		Msg("twilio webhook: recipient status updated")

	c.Status(http.StatusOK)
}

// ----- SendGrid -----

// sendGridEvent represents a single event in the SendGrid event webhook payload.
type sendGridEvent struct {
	SGMessageID string `json:"sg_message_id"`
	Event       string `json:"event"`
	Email       string `json:"email"`
	Timestamp   int64  `json:"timestamp"`
	Reason      string `json:"reason,omitempty"`
}

// sendGridEventMap maps SendGrid event types to internal MessageStatus values.
var sendGridEventMap = map[string]model.MessageStatus{
	"delivered": model.StatusDelivered,
	"bounce":    model.StatusFailed,
	"dropped":   model.StatusFailed,
}

// SendGridWebhook handles POST /webhooks/sendgrid
// SendGrid sends a JSON array of event objects.
func (h *WebhookHandler) SendGridWebhook(c *gin.Context) {
	var events []sendGridEvent
	if err := c.ShouldBindJSON(&events); err != nil {
		logger.Get().Error().Err(err).Msg("sendgrid webhook: failed to parse payload")
		c.Status(http.StatusBadRequest)
		return
	}

	for _, evt := range events {
		internalStatus, ok := sendGridEventMap[strings.ToLower(evt.Event)]
		if !ok {
			// Unhandled event type (e.g., open, click) — skip
			continue
		}

		// SendGrid message IDs may have a trailing ".filter..." suffix — strip it
		msgID := cleanSendGridMessageID(evt.SGMessageID)

		recipient, err := h.recipientRepo.GetByProviderID(c.Request.Context(), msgID)
		if err != nil {
			if err == repository.ErrNotFound {
				logger.Get().Warn().Str("sg_message_id", msgID).Msg("sendgrid webhook: recipient not found")
				continue
			}
			logger.Get().Error().Err(err).Str("sg_message_id", msgID).Msg("sendgrid webhook: lookup failed")
			continue
		}

		if err := h.recipientRepo.UpdateStatus(c.Request.Context(), recipient.ID, internalStatus, &msgID); err != nil {
			logger.Get().Error().Err(err).Str("sg_message_id", msgID).Msg("sendgrid webhook: failed to update status")
			continue
		}

		logger.Get().Info().
			Str("sg_message_id", msgID).
			Str("event", evt.Event).
			Str("recipient_id", recipient.ID.String()).
			Msg("sendgrid webhook: recipient status updated")
	}

	c.Status(http.StatusOK)
}

// cleanSendGridMessageID removes the ".filter..." suffix that SendGrid
// sometimes appends to sg_message_id values.
func cleanSendGridMessageID(id string) string {
	if idx := strings.Index(id, ".filter"); idx != -1 {
		return id[:idx]
	}
	return id
}
