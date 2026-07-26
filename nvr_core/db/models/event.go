package models

import (
	"time"
)

// SystemLicense represents a database record in the system_licenses table.
type Event struct {
	ID         int64     `json:"id" db:"id"`
	Type       string    `json:"type" db:"type"`
	Message    string    `json:"message" db:"message"`
	CreatedAt  time.Time `json:"created_at" db:"create_datetime"`
}
