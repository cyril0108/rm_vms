package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type EventType string

const (
	EventTypeMotion         EventType = "MOTION_DETECTED"
	EventTypeCameraOffline  EventType = "CAMERA_OFFLINE"
	EventTypeDiskWarning    EventType = "DISK_WARNING"
)

type EventPayload map[string]interface{}

type Event struct {
	ID        int64        `json:"id" db:"id"`
	Type      EventType    `json:"type" db:"type"`
	Message   string       `json:"message" db:"message"`
	Payload   EventPayload `json:"payload" db:"payload"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
}

// Value makes EventPayload satisfy the driver.Valuer interface.
// It automatically converts the map to a JSON string when saving to the DB.
func (p EventPayload) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan makes EventPayload satisfy the sql.Scanner interface.
// It automatically converts the JSON string from SQLite back into a Go map.
func (p *EventPayload) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return errors.New("type assertion to []byte failed in EventPayload")
	}

	return json.Unmarshal(bytes, p)
}
