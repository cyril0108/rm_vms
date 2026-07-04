package utils

import (
	"crypto/rand"
	"encoding/hex"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// GenerateTaskID creates a cryptographically secure random string
func GenerateTaskID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback in the incredibly rare event the system entropy pool is exhausted
		return "fallback_id" 
	}
	return hex.EncodeToString(bytes)
}