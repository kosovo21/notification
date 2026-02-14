package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"notification-system/internal/auth"
	"notification-system/internal/config"
	"notification-system/internal/repository"
)

func main() {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to DB
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	// Generate API Key
	apiKey := generateRandomKey(32)
	apiKeyHash := auth.HashAPIKey(apiKey)

	// User data
	userID := uuid.New()
	email := fmt.Sprintf("test-%s@example.com", apiKey[:8])

	// Insert User
	query := `
		INSERT INTO users (id, email, api_key_hash, role, rate_limit_tier, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, 'user', 'premium', true, NOW(), NOW())
		ON CONFLICT (email) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`

	var id uuid.UUID
	err = db.QueryRow(query, userID, email, apiKeyHash).Scan(&id)
	if err != nil {
		log.Fatalf("failed to seed user: %v", err)
	}

	fmt.Println("---------------------------------------------------")
	fmt.Println("✅  Test User Created Successfully")
	fmt.Println("---------------------------------------------------")
	fmt.Println("User ID:   ", id)
	fmt.Println("Email:     ", email)
	fmt.Println("API Key:   ", apiKey)
	fmt.Println("---------------------------------------------------")
	fmt.Println("Use this API Key in the 'X-API-Key' header")
	fmt.Println("---------------------------------------------------")
}

func generateRandomKey(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(bytes)
}
