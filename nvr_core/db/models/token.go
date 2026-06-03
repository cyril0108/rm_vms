package models

import "time"

type RefreshToken struct {
	ID        string    // UUID string
	UserID    int64     // Matches the int64 ID from your User model
	TokenHash string    // The bcrypt hash of the raw token
	UserAgent string    // Optional: Browser/Device info
	ClientIP  string    // Optional: IP Address
	IsRevoked bool      // Soft-delete flag for audit trails
	ExpiresAt time.Time // When the token natively dies
	CreatedAt time.Time
}