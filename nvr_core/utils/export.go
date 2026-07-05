package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusFinalizing TaskStatus = "finalizing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// SupportedExportFormats is the single source of truth for NVR video export formats and their MIME types.
var SupportedExportFormats = map[string]string{
	"mp4":  "video/mp4",
	"mkv":  "video/x-matroska",
	"avi":  "video/x-msvideo",
	"mov":  "video/quicktime",
	"webm": "video/webm",
}

// GenerateTaskID creates a cryptographically secure random string
func GenerateTaskID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback in the incredibly rare event the system entropy pool is exhausted
		return "fallback_id" 
	}
	return hex.EncodeToString(bytes)
}

// GetVideoMimeType safely looks up the MIME type based on the file path.
func GetVideoMimeType(ext string) string {

	if mimeType, exists := SupportedExportFormats[ext]; exists {
		return mimeType
	}

	// Safe fallback for generic binary data
	return "application/octet-stream"
}

func ValidateExportFormat(format string) error {
	format = strings.ToLower(format)

	// Check if the format exists in our central map
	if _, exists := SupportedExportFormats[format]; !exists {

		// Dynamically build the allowed formats string for the error message
		allowed := make([]string, 0, len(SupportedExportFormats))
		for key := range SupportedExportFormats {
			allowed = append(allowed, key)
		}

		return fmt.Errorf("unsupported export format: %s. Allowed formats: %s", format, strings.Join(allowed, ", "))
	}

	return nil
}