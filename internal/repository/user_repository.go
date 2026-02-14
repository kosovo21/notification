package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"notification-system/internal/model"
)

// ErrNotFound is returned when a query finds no matching rows.
var ErrNotFound = errors.New("record not found")

// UserRepository defines data access operations for users.
type UserRepository interface {
	GetByAPIKeyHash(ctx context.Context, hash string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository backed by sqlx.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByAPIKeyHash(ctx context.Context, hash string) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, api_key_hash, role, rate_limit_tier, is_active, created_at, updated_at
	           FROM users WHERE api_key_hash = $1`

	if err := r.db.GetContext(ctx, &user, query, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, api_key_hash, role, rate_limit_tier, is_active, created_at, updated_at
	           FROM users WHERE id = $1`

	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
