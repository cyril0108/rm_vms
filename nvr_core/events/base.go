package events

import (
	"context"
	// "sync"
	// "time"

	"nvr_core/db/models"
	"nvr_core/logger"
)

var LOG = logger.NewLogger("\033[3m[nvr_core][events]\033[0m")


// Notifier is the interface that Email, Push, and WebSockets will implement
type Notifier interface {
	// ID returns the name of the notifier (e.g., "EmailService")
	ID() string 
	// Handle asynchronously processes the event
	Handle(ctx context.Context, e models.Event)
}
