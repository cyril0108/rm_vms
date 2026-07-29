package models

import (
	"time"
)

type EventType string

const (
	EventTypeMotion         EventType = "MOTION_DETECTED"
	EventTypeCameraOffline  EventType = "CAMERA_OFFLINE"
	EventTypeDiskWarning    EventType = "DISK_WARNING"
)

type EventPayload map[string]interface{}

// SystemLicense represents a database record in the system_licenses table.
type Event struct {
	ID        int64        `json:"id" db:"id"`
	Type      EventType    `json:"type" db:"type"`
	Message   string       `json:"message" db:"message"`
	Payload   EventPayload `json:"payload" db:"payload"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
}
