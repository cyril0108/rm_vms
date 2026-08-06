package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"nvr_core/db/models"
)

/**
 * Event Basic Operations
 */

// EventRepository defines the contract for segment data access.
type EventRepository interface {

	Insert(ctx context.Context, ev *models.Event) error
	GetLastestEvents(ctx context.Context, limit int) ([]*models.Event, error)
	GetLastEvent(ctx context.Context) (*models.Event, error)
	GetEventsFrom(ctx context.Context, time time.Time) ([]*models.Event, error)
	GetEventsBetween(ctx context.Context, start, end time.Time) ([]*models.Event, error)

}

type eventRepo struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepo{db: db}
}

func (e *eventRepo) Insert(ctx context.Context, ev *models.Event) error {
	query := `INSERT INTO events (type, message, payload, created_at) 
	          VALUES (?, ?, ?, ?)`

	result, err := e.db.ExecContext(ctx, query, ev.Type, ev.Message, ev.Payload, ev.CreatedAt)
	if err != nil {
		return err
	}
	ev.ID, _ = result.LastInsertId()
	return nil
}

func (e *eventRepo) GetLastEvent(ctx context.Context) (*models.Event, error) {
	query := `
		SELECT id, type, message, created_at
		FROM events
		ORDER BY created_at DESC 
		LIMIT 1
	`

	var ev models.Event
	err := e.db.QueryRowContext(ctx, query).Scan(
		&ev.ID,
		&ev.Type,
		&ev.Message,
		&ev.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil gracefully if the database is completely empty
		}
		return nil, err
	}

	return &ev, nil
}

func (e *eventRepo) scanRows(rows *sql.Rows) ([]*models.Event, error) {

	var events []*models.Event
	for rows.Next() {
		var ev models.Event
		if err := rows.Scan( &ev.ID, &ev.Type, &ev.Message, &ev.Payload, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &ev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil

}

func (e *eventRepo) GetLastestEvents(ctx context.Context, limit int) ([]*models.Event, error) {

	query := `
		SELECT id, type, message, payload, created_at
		FROM events
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := e.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return e.scanRows(rows)

}

func (e *eventRepo) GetEventsBetween(ctx context.Context, start, end time.Time) ([]*models.Event, error) {

	query := `
		SELECT id, type, message, payload, created_at
		FROM events
		WHERE
			created_at > ?
			AND
			created_at < ?
		ORDER BY created_at DESC
	`

	rows, err := e.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return e.scanRows(rows)

}

func (e *eventRepo) GetEventsFrom(ctx context.Context, time time.Time) ([]*models.Event, error) {
	query := `
		SELECT id, type, message, payload, created_at
		FROM events
		WHERE
			created_at > ?
		ORDER BY created_at DESC
	`

	rows, err := e.db.QueryContext(ctx, query, time)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return e.scanRows(rows)

}
