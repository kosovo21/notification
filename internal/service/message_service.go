package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"notification-system/internal/model"
	"notification-system/internal/queue"
	"notification-system/internal/repository"
)

// MessageService handles message processing logic.
type MessageService struct {
	db            *sqlx.DB
	messageRepo   repository.MessageRepository
	recipientRepo repository.RecipientRepository
	publisher     *queue.Publisher
}

// NewMessageService creates a new MessageService.
func NewMessageService(
	db *sqlx.DB,
	messageRepo repository.MessageRepository,
	recipientRepo repository.RecipientRepository,
	publisher *queue.Publisher,
) *MessageService {
	return &MessageService{
		db:            db,
		messageRepo:   messageRepo,
		recipientRepo: recipientRepo,
		publisher:     publisher,
	}
}

// SendMessage handles the creation and queuing of a message.
func (s *MessageService) SendMessage(ctx context.Context, userID uuid.UUID, req model.CreateMessageRequest) (*model.SendMessageResponse, error) {
	now := time.Now()
	msgID := uuid.New()

	priority := model.PriorityNormal
	if req.Priority != nil {
		priority = model.Priority(*req.Priority)
	}

	// Scheduled messages are saved but not published until the scheduler picks them up
	isScheduled := req.ScheduledAt != nil && req.ScheduledAt.After(now)
	status := model.StatusPending
	if isScheduled {
		status = model.StatusScheduled
	}

	msg := &model.Message{
		ID:          msgID,
		UserID:      userID,
		Subject:     req.Subject,
		Body:        req.Message,
		Sender:      req.From,
		Platform:    model.Platform(req.Platform),
		Priority:    priority,
		Status:      status,
		ScheduledAt: req.ScheduledAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	recipients := make([]model.Recipient, len(req.To))
	for i, to := range req.To {
		recipients[i] = model.Recipient{
			ID:         uuid.New(),
			MessageID:  msgID,
			Recipient:  to,
			Status:     model.StatusPending,
			RetryCount: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.messageRepo.Create(ctx, tx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	if err := s.recipientRepo.BatchCreate(ctx, tx, recipients); err != nil {
		return nil, fmt.Errorf("failed to create recipients: %w", err)
	}

	// Only publish immediately if not scheduled
	if !isScheduled {
		if err := s.publishRecipients(ctx, msg, recipients); err != nil {
			return nil, err
		}

		_, err := tx.ExecContext(ctx,
			"UPDATE messages SET status = $1, updated_at = $2 WHERE id = $3",
			model.StatusQueued, time.Now(), msgID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update message status: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &model.SendMessageResponse{
		Success:           true,
		MessageID:         msgID.String(),
		RecipientsCount:   len(recipients),
		EstimatedDelivery: now.Add(30 * time.Second),
		RequestID:         uuid.New().String(),
	}, nil
}

// publishRecipients fans out events to RabbitMQ for each recipient.
func (s *MessageService) publishRecipients(ctx context.Context, msg *model.Message, recipients []model.Recipient) error {
	routingKey := platformToRoutingKey(msg.Platform)

	for _, r := range recipients {
		event := queue.MessageQueuedEvent{
			MessageID:   msg.ID.String(),
			RecipientID: r.ID.String(),
			To:          r.Recipient,
			Body:        msg.Body,
			Subject:     msg.Subject,
			Platform:    string(msg.Platform),
			Timestamp:   time.Now(),
		}

		if err := s.publisher.Publish(ctx, queue.ExchangeName, routingKey, event); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	return nil
}

// PublishMessage publishes an already-persisted message's recipients to RabbitMQ.
// Used by the scheduler to dispatch scheduled messages.
func (s *MessageService) PublishMessage(ctx context.Context, msg *model.Message) error {
	recipients, err := s.recipientRepo.GetByMessageID(ctx, msg.ID)
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}

	if err := s.publishRecipients(ctx, msg, recipients); err != nil {
		return err
	}

	if err := s.messageRepo.UpdateStatus(ctx, msg.ID, model.StatusQueued); err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	log.Info().
		Str("message_id", msg.ID.String()).
		Int("recipients", len(recipients)).
		Msg("scheduled message published")

	return nil
}

// platformToRoutingKey maps a platform to its RabbitMQ routing key.
func platformToRoutingKey(p model.Platform) string {
	switch p {
	case model.PlatformSMS:
		return queue.RoutingKeySMS
	case model.PlatformEmail:
		return queue.RoutingKeyEmail
	case model.PlatformWhatsApp:
		return queue.RoutingKeyWhatsApp
	case model.PlatformTelegram:
		return queue.RoutingKeyTelegram
	default:
		return queue.RoutingKeyEmail
	}
}
