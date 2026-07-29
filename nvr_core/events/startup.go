package events

import (
	"context"
	"database/sql"
	"nvr_core/db/repository"
)


func StartUp(ctx context.Context, db *sql.DB) *EventHub {

	eventRepo := repository.NewEventRepository(db)

	hub := NewEventHub(500)

	hub.RegisterNotifier(NewDatabaseNotifier(eventRepo))
	// hub.RegisterNotifier(events.NewWebSocketNotifier(wsHub))
	// hub.RegisterNotifier(events.NewEmailNotifier(emailConfig))

	// Start the background dispatcher
	go hub.Start(ctx)

	return &hub
}