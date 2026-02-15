package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"notification-system/internal/auth"
	"notification-system/internal/config"
	"notification-system/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Connect to database
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	ctx := context.Background()

	// Define test user
	apiKey := "test-api-key"
	hashedKey := auth.HashAPIKey(apiKey)
	userID := uuid.New()
	email := "test@example.com"

	// Check if user exists
	query := `INSERT INTO users (id, email, api_key_hash, role, rate_limit_tier, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, 'admin', 'premium', true, NOW(), NOW())
	          ON CONFLICT (email) DO UPDATE SET api_key_hash = EXCLUDED.api_key_hash`

	_, err = db.ExecContext(ctx, query, userID, email, hashedKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to seed user")
	}

	logger.Info().Msgf("Seeded user: %s", email)
	logger.Info().Msgf("API Key: %s", apiKey)
}
