package events

import (
	"context"
	"sync"
	"time"

	"nvr_core/db/models"
)

type EventHub interface {
	SendEvent(etype models.EventType, message string, payload *models.EventPayload)
	Publish(e models.Event)
	RegisterNotifier(n Notifier)
	Start(ctx context.Context)
}

type eventService struct {
	mu         sync.RWMutex
	eventChan  chan models.Event
	notifiers  []Notifier
}

// NewEventHub initializes the hub with a buffered channel
func NewEventHub(bufferSize int) EventHub {
	return &eventService{
		eventChan: make(chan models.Event, bufferSize),
		notifiers: make([]Notifier, 0),
	}
}

// RegisterNotifier allows different modules to subscribe on boot
func (s *eventService) RegisterNotifier(n Notifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifiers = append(s.notifiers, n)
	LOG.Info("[EventService] Registered Notifier", "notifier", n.ID())
}

func (s *eventService) SendEvent(etype models.EventType, message string, payload *models.EventPayload) {
	e := models.Event {
		Type: etype,
		Message: message,
		CreatedAt: time.Now(),
	}
	if payload != nil {
		e.Payload = *payload
	}
	s.Publish(e)
}


// Publish safely injects an event into the channel without blocking
func (s *eventService) Publish(e models.Event) {
	// CRITICAL: We use a non-blocking select. 
	// If the event channel is completely full, we drop the event 
	// rather than freezing the caller (like a live video ingestion thread).
	select {
	case s.eventChan <- e:
		// Successfully queued
	default:
		LOG.Info("[EventService] WARNING: Event channel full! Dropped event", "type", e.Type, "message", e.Message)
	}
}


func (s *eventService) Start(ctx context.Context) {
	LOG.Info("[EventService] Starting event dispatcher...")

	for {
		select {
		case <-ctx.Done():
			LOG.Info("[EventService] Shutting down.")
			return

		case e := <-s.eventChan:
			s.mu.RLock()
			// Fan-out the event to all registered notifiers
			for _, notifier := range s.notifiers {
				// We wrap the execution in a goroutine. 
				// If the Email service takes 2 seconds to connect to SMTP, 
				// it will NOT block the WebSocket service from sending instantly.
				go func(n Notifier, evt models.Event) {
					// Use a timeout context so a dead API doesn't leak goroutines
					timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()

					n.Handle(timeoutCtx, evt)
				}(notifier, e)
			}
			s.mu.RUnlock()
		}
	}
}