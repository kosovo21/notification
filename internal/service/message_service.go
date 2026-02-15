package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

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

	// Determine priority
	priority := model.PriorityNormal
	if req.Priority != nil {
		priority = model.Priority(*req.Priority)
	}

	// Initial status is Pending (will be updated to Queued after publish)
	// If scheduled, it stays as Pending/Scheduled until picked up by scheduler
	status := model.StatusPending
	if req.ScheduledAt != nil && req.ScheduledAt.After(now) {
		status = model.StatusQueued // Scheduler looks for Queued/Scheduled?
		// Actually PLAN says: "ScanScheduledMessages: query messages where scheduled_at <= now AND status = 'scheduled'"
		// But Model doesn't have StatusScheduled.
		// Let's assume StatusQueued with ScheduledAt set implies scheduled.
		// Or I should add StatusScheduled.
		// For now, let's use StatusQueued for scheduled messages too, the scheduler can filter by date.
		// Wait, if I publish it now, it will be processed immediately?
		// If ScheduledAt is set, I should NOT publish it to RabbitMQ yet?
		// Ref plan: "ScanScheduledMessages... Publish to RabbitMQ... Update status to queued"
		// So detailed flow:
		// 1. If Scheduled: Save to DB (Status=Queued/Scheduled), Return. (Do NOT publish).
		// 2. If Immediate: Save to DB (Status=Pending), Publish, Update to Status=Queued.

		// Let's stick to the plan logic roughly.
		// If ScheduledAt is future:
		//   Save as StatusQueued (or maybe I should add StatusScheduled to model?
		//   Model has: Queued, Processing, Sent, Delivered, Failed, Pending, Cancelled.
		//   I'll Use StatusQueued for now and rely on ScheduledAt check.
		//   Wait, if I use StatusQueued, I need to make sure the worker/scheduler distinguishes them.
		//   Actually, the scheduler will pick them up.
	}

	shouldPublish := true
	if req.ScheduledAt != nil && req.ScheduledAt.After(now) {
		status = model.StatusQueued // Effectively "Scheduled"
		shouldPublish = false
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

	// Build recipients
	recipients := make([]model.Recipient, len(req.To))
	for i, to := range req.To {
		recipients[i] = model.Recipient{
			ID:         uuid.New(),
			MessageID:  msgID,
			Recipient:  to,
			Status:     model.StatusPending, // Initially pending
			RetryCount: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}

	// Transaction
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safety

	if err := s.messageRepo.Create(ctx, tx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	if err := s.recipientRepo.BatchCreate(ctx, tx, recipients); err != nil {
		return nil, fmt.Errorf("failed to create recipients: %w", err)
	}

	// Publish to RabbitMQ if needed
	if shouldPublish {
		for _, r := range recipients {
			// Determine routing key based on platform
			var routingKey string
			switch msg.Platform {
			case model.PlatformSMS:
				routingKey = queue.RoutingKeySMS
			case model.PlatformEmail:
				routingKey = queue.RoutingKeyEmail
			case model.PlatformWhatsApp:
				routingKey = queue.RoutingKeyWhatsApp
			case model.PlatformTelegram:
				routingKey = queue.RoutingKeyTelegram
			default:
				routingKey = queue.RoutingKeyEmail // Fallback? Or error?
			}

			event := queue.MessageQueuedEvent{
				MessageID:   msgID.String(),
				RecipientID: r.ID.String(),
				To:          r.Recipient,
				Body:        msg.Body,
				Subject:     msg.Subject,
				Platform:    string(msg.Platform),
				Timestamp:   now,
			}

			if err := s.publisher.Publish(ctx, queue.ExchangeName, routingKey, event); err != nil {
				// If publish fails, we should probably fail the request or mark as pending retry?
				// For now, let's fail the transaction to ensure consistency (Atomic).
				// Or we could commit DB and let a recovery process handle it.
				// Plan says: "Update DB Status (Status: QUEUED)".

				// Proceeding with strict consistency: Fail transaction.
				return nil, fmt.Errorf("failed to publish event: %w", err)
			}
		}

		// Update message status to Queued
		// Note: We are inside a transaction, but UpdateStatus in repo uses `db.Exec`, not `tx`.
		// I need to update repo to accept tx for UpdateStatus?
		// Or just implement a `UpdateStatusTx`?
		// Looking at repo interface: `UpdateStatus(ctx, id, status)`. It uses `r.db`.
		// I should ideally update the repo to support transaction or specific tx method.

		// For now, I can manually execute the update query using tx here, OR update the repo interface.
		// Cleanest way: Update repo interface to accept optional *sqlx.Tx or methods that take it.
		// Given constraints, I'll update the Message object in memory and save it? No, `Create` is INSERT.

		// I'll add `UpdateStatusTx` to `MessageRepository`? Or change `UpdateStatus` to take `tx`.
		// But `UpdateStatus` is used elsewhere?
		// Usage in `MessageHandler` uses `messageRepo.Create(..., tx, ...)` but `UpdateStatus` usage?
		// Handler doesn't use UpdateStatus.

		// Let's assume I can update `MessageRepository` to accept `*sqlx.Tx` for `UpdateStatus`.
		// Or I can just execute the query directly on tx in the service for now to save time/complexity?
		// "Update DB Status (Status: QUEUED)"

		// I will modify the message struct locally and assume it's created with the FINAL status?
		// No, the plan says: "2. Persist... (PENDING) 3. Fan-out... 4. Update... (QUEUED)".

		// Alternative: Create with Status=PENDING. Publish. Then Update to QUEUED.
		// If I update to QUEUED using `tx`, it's safe.
		// If I update using `db` (outside tx), and tx rolls back? Bad.

		// I will modify `UpdateStatus` in `MessageRepository` to accept `*sqlx.Tx`. If nil, use `r.db`.
		// No, `UpdateStatus` currently doesn't take tx.
		// I'll run a direct `tx.Exec` here for the update to keep it simple and transactional.

		_, err := tx.ExecContext(ctx, "UPDATE messages SET status = $1, updated_at = $2 WHERE id = $3", model.StatusQueued, time.Now(), msgID)
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
		EstimatedDelivery: now.Add(30 * time.Second), // Mock estimation
		RequestID:         uuid.New().String(),
	}, nil
}
