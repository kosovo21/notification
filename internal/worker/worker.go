package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"notification-system/internal/adapter"
	"notification-system/internal/model"
	"notification-system/internal/queue"
	"notification-system/internal/repository"
)

// Worker processes queued notification events.
type Worker struct {
	consumer      *queue.Consumer
	recipientRepo repository.RecipientRepository
	messageRepo   repository.MessageRepository
	adapters      map[string]adapter.Sender
}

// NewWorker creates a new Worker.
func NewWorker(
	consumer *queue.Consumer,
	recipientRepo repository.RecipientRepository,
	messageRepo repository.MessageRepository,
	adapters map[string]adapter.Sender,
) *Worker {
	return &Worker{
		consumer:      consumer,
		recipientRepo: recipientRepo,
		messageRepo:   messageRepo,
		adapters:      adapters,
	}
}

// Start begins consuming messages from the given queue.
func (w *Worker) Start(ctx context.Context, queueName, routingKey string) error {
	log.Info().
		Str("queue", queueName).
		Str("routing_key", routingKey).
		Msg("worker started")

	return w.consumer.Consume(ctx, queueName, routingKey, w.processMessage)
}

// processMessage handles a single queued message event.
func (w *Worker) processMessage(ctx context.Context, body []byte) error {
	var event queue.MessageQueuedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal event")
		return fmt.Errorf("unmarshal error: %w", err)
	}

	log.Info().
		Str("message_id", event.MessageID).
		Str("recipient_id", event.RecipientID).
		Str("platform", event.Platform).
		Str("to", event.To).
		Msg("processing notification")

	recipientID, err := uuid.Parse(event.RecipientID)
	if err != nil {
		return fmt.Errorf("invalid recipient ID: %w", err)
	}

	// Update recipient status to Processing
	if err := w.recipientRepo.UpdateStatus(ctx, recipientID, model.StatusProcessing, nil); err != nil {
		log.Error().Err(err).Str("recipient_id", event.RecipientID).Msg("failed to update recipient status to processing")
		// Continue processing anyway
	}

	// Select adapter
	senderAdapter, ok := w.adapters[event.Platform]
	if !ok {
		errMsg := fmt.Sprintf("no adapter for platform: %s", event.Platform)
		log.Error().Str("platform", event.Platform).Msg(errMsg)
		w.recipientRepo.UpdateStatus(ctx, recipientID, model.StatusFailed, nil)
		return fmt.Errorf(errMsg)
	}

	// Send notification
	result, err := senderAdapter.Send(ctx, event.To, event.Subject, event.Body)
	if err != nil {
		log.Error().Err(err).
			Str("message_id", event.MessageID).
			Str("to", event.To).
			Msg("failed to send notification")

		w.recipientRepo.UpdateStatus(ctx, recipientID, model.StatusFailed, nil)
		return fmt.Errorf("send failed: %w", err)
	}

	// Update recipient status to Sent
	if err := w.recipientRepo.UpdateStatus(ctx, recipientID, model.StatusSent, &result.ProviderID); err != nil {
		log.Error().Err(err).Str("recipient_id", event.RecipientID).Msg("failed to update recipient status to sent")
		return fmt.Errorf("status update error: %w", err)
	}

	log.Info().
		Str("message_id", event.MessageID).
		Str("recipient_id", event.RecipientID).
		Str("provider_id", result.ProviderID).
		Msg("notification sent successfully")

	return nil
}
