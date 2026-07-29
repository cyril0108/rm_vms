package service

import (
	"context"
	"time"

	"nvr_core/db/models"
	"nvr_core/db/repository"
)


type EventService interface {
	NewEvent(ctx context.Context, evType models.EventType, message string) (*models.Event, error)
	GetEventsFrom(ctx context.Context, time time.Time) ([]*models.Event, error)
}

func NewEventService(repo repository.EventRepository) EventService {
	return &eventServiceBase{repo: repo}
}

func (e *eventServiceBase) NewEvent(ctx context.Context, evType models.EventType, message string) (*models.Event, error) {
	event := &models.Event{
		Type: evType,
		Message: message,
		CreatedAt: time.Now(),
	}
	if err := e.repo.Insert(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (e *eventServiceBase) GetEventsFrom(ctx context.Context, time time.Time) ([]*models.Event, error) {
	return e.repo.GetEventsFrom(ctx, time)
}

func (e *eventServiceBase) GetLastEvent(ctx context.Context) (*models.Event, error) {
	return e.repo.GetLastEvent(ctx)
}
