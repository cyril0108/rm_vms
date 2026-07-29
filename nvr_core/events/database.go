package events

import (
	"context"
	"nvr_core/db/models"
	"nvr_core/db/repository"
)

var log = LOG.Prefix("[database]")

// DatabaseNotifier listens to the EventHub and persists events to SQLite
type DatabaseNotifier struct {
	repo repository.EventRepository
}

// NewDatabaseNotifier creates a new instance of the notifier
func NewDatabaseNotifier(repo repository.EventRepository) Notifier {
	return &DatabaseNotifier{
		repo: repo,
	}
}

// ID satisfies the Notifier interface
func (d *DatabaseNotifier) ID() string {
	return "DatabaseNotifier"
}

// Handle receives the broadcasted event and writes it to the database
func (d *DatabaseNotifier) Handle(ctx context.Context, e models.Event) {
	// The EventHub has already spawned this in a background goroutine with a timeout context.
	// We can safely perform our synchronous database insertion here.
	err := d.repo.Insert(ctx, &e)
	if err != nil {
		// Log the error, but know that it won't crash the EventHub or other notifiers
		log.Error("[DatabaseNotifier] Failed to insert event to database", "type", e.Type, "error", err)
	}
}