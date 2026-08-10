package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type EmailSMTPSettings struct {
	ID          int64     `json:"id" db:"id"`
	Host        string    `json:"host" db:"host"`
	Port        int       `json:"port" db:"port"`
	Username    string    `json:"username" db:"username"`
	Password    string    `json:"password" db:"password"`
	SenderEmail string    `json:"sender_email" db:"sender_email"`
	SenderName  string    `json:"sender_name" db:"sender_name"`
	UseTLS      bool      `json:"use_tls" db:"use_tls"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// EmailRecipients is a JSON-serializable string slice stored as TEXT in SQLite.
type EmailRecipients []string

type EmailGroup struct {
	ID         int64           `json:"id" db:"id"`
	Name       string          `json:"name" db:"name"`
	Recipients EmailRecipients `json:"recipients" db:"recipients"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

type EmailGroupEvent struct {
	ID        int64  `json:"id" db:"id"`
	GroupID   int64  `json:"group_id" db:"group_id"`
	EventType string `json:"event_type" db:"event_type"`
}

// Value makes EmailRecipients satisfy the driver.Valuer interface.
// It converts the string slice to a JSON array when saving to the DB.
func (r EmailRecipients) Value() (driver.Value, error) {
	if r == nil {
		r = []string{}
	}
	return json.Marshal(r)
}

// Scan makes EmailRecipients satisfy the sql.Scanner interface.
// It converts the JSON array from SQLite back into a Go string slice.
func (r *EmailRecipients) Scan(value interface{}) error {
	if value == nil {
		*r = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return errors.New("type assertion to []byte failed in EmailRecipients")
	}

	return json.Unmarshal(bytes, r)
}
