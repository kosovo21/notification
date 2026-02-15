package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"notification-system/internal/model"
)

// RecipientRepository defines data access operations for message recipients.
type RecipientRepository interface {
	BatchCreate(ctx context.Context, tx *sqlx.Tx, recipients []model.Recipient) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.MessageStatus, providerID *string) error
	GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]model.Recipient, error)
	GetByProviderID(ctx context.Context, providerID string) (*model.Recipient, error)
}

type recipientRepository struct {
	db *sqlx.DB
}

// NewRecipientRepository creates a new RecipientRepository backed by sqlx.
func NewRecipientRepository(db *sqlx.DB) RecipientRepository {
	return &recipientRepository{db: db}
}

func (r *recipientRepository) BatchCreate(ctx context.Context, tx *sqlx.Tx, recipients []model.Recipient) error {
	query := `INSERT INTO message_recipients (id, message_id, recipient, status, retry_count, created_at, updated_at)
	           VALUES (:id, :message_id, :recipient, :status, :retry_count, :created_at, :updated_at)`

	_, err := tx.NamedExecContext(ctx, query, recipients)
	return err
}

func (r *recipientRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.MessageStatus, providerID *string) error {
	now := time.Now()

	query := `UPDATE message_recipients
	           SET status = $1, provider_id = $2, updated_at = $3`

	// Set timestamp columns based on status
	switch status {
	case model.StatusSent:
		query += `, sent_at = $4 WHERE id = $5`
		result, err := r.db.ExecContext(ctx, query, status, providerID, now, now, id)
		if err != nil {
			return err
		}
		return checkRowsAffected(result)
	case model.StatusDelivered:
		query += `, delivered_at = $4 WHERE id = $5`
		result, err := r.db.ExecContext(ctx, query, status, providerID, now, now, id)
		if err != nil {
			return err
		}
		return checkRowsAffected(result)
	default:
		query += ` WHERE id = $4`
		result, err := r.db.ExecContext(ctx, query, status, providerID, now, id)
		if err != nil {
			return err
		}
		return checkRowsAffected(result)
	}
}

func (r *recipientRepository) GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]model.Recipient, error) {
	var recipients []model.Recipient
	query := `SELECT id, message_id, recipient, status, provider_id, error_message, retry_count,
	                  sent_at, delivered_at, created_at, updated_at
	           FROM message_recipients WHERE message_id = $1 ORDER BY created_at`

	if err := r.db.SelectContext(ctx, &recipients, query, messageID); err != nil {
		return nil, err
	}

	return recipients, nil
}

func (r *recipientRepository) GetByProviderID(ctx context.Context, providerID string) (*model.Recipient, error) {
	var recipient model.Recipient
	query := `SELECT id, message_id, recipient, status, provider_id, error_message, retry_count,
	                  sent_at, delivered_at, created_at, updated_at
	           FROM message_recipients WHERE provider_id = $1`

	if err := r.db.GetContext(ctx, &recipient, query, providerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &recipient, nil
}

// checkRowsAffected returns ErrNotFound if no rows were updated.
func checkRowsAffected(result interface{ RowsAffected() (int64, error) }) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
