package models

import "time"

// User represents a system operator or viewer in the NVR.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Password     string    `json:"-"` // CRITICAL: Never serialize to JSON
	RoleID       int64     `json:"role_id"`
	Email        string    `json:"email"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}