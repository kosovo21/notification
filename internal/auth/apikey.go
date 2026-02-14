package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashAPIKey returns the SHA-256 hex digest of a raw API key.
// This is used to compare against the api_key_hash column in the DB.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
