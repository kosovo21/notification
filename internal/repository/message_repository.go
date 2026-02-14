package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"notification-system/internal/model"
)

// MessageRepository defines data access operations for messages.
type MessageRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, msg *model.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Message, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.MessageStatus) error
	List(ctx context.Context, userID uuid.UUID, q model.ListMessagesQuery) ([]model.Message, int, error)
}

type messageRepository struct {
	db *sqlx.DB
}

// NewMessageRepository creates a new MessageRepository backed by sqlx.
func NewMessageRepository(db *sqlx.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, tx *sqlx.Tx, msg *model.Message) error {
	query := `INSERT INTO messages (id, user_id, subject, body, sender, platform, priority, status, scheduled_at, created_at, updated_at)
	           VALUES (:id, :user_id, :subject, :body, :sender, :platform, :priority, :status, :scheduled_at, :created_at, :updated_at)`

	_, err := tx.NamedExecContext(ctx, query, msg)
	return err
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Message, error) {
	var msg model.Message
	query := `SELECT id, user_id, subject, body, sender, platform, priority, status, scheduled_at, created_at, updated_at
	           FROM messages WHERE id = $1`

	if err := r.db.GetContext(ctx, &msg, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &msg, nil
}

func (r *messageRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.MessageStatus) error {
	query := `UPDATE messages SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *messageRepository) List(ctx context.Context, userID uuid.UUID, q model.ListMessagesQuery) ([]model.Message, int, error) {
	// Build dynamic WHERE clause
	conditions := []string{"user_id = :user_id"}
	params := map[string]interface{}{
		"user_id": userID,
	}

	if q.Platform != "" {
		conditions = append(conditions, "platform = :platform")
		params["platform"] = q.Platform
	}
	if q.Status != nil {
		conditions = append(conditions, "status = :status")
		params["status"] = *q.Status
	}
	if q.From != nil {
		conditions = append(conditions, "created_at >= :from_date")
		params["from_date"] = *q.From
	}
	if q.To != nil {
		conditions = append(conditions, "created_at <= :to_date")
		params["to_date"] = *q.To
	}

	where := strings.Join(conditions, " AND ")

	// Count total matching rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM messages WHERE %s", where)
	countQuery, countArgs, err := sqlx.Named(countQuery, params)
	if err != nil {
		return nil, 0, err
	}
	countQuery = r.db.Rebind(countQuery)

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	offset := (q.Page - 1) * q.Limit
	params["limit"] = q.Limit
	params["offset"] = offset

	dataQuery := fmt.Sprintf(
		`SELECT id, user_id, subject, body, sender, platform, priority, status, scheduled_at, created_at, updated_at
		 FROM messages WHERE %s ORDER BY created_at DESC LIMIT :limit OFFSET :offset`, where)

	dataQuery, dataArgs, err := sqlx.Named(dataQuery, params)
	if err != nil {
		return nil, 0, err
	}
	dataQuery = r.db.Rebind(dataQuery)

	var messages []model.Message
	if err := r.db.SelectContext(ctx, &messages, dataQuery, dataArgs...); err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}
