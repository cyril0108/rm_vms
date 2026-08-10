package service

import (
	"context"

	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/email"
)

// EmailGroupWithEvents bundles a group with its subscribed event types.
type EmailGroupWithEvents struct {
	Group      models.EmailGroup
	EventTypes []string
}

type EmailService interface {
	// SMTP Settings
	GetSMTPSettings(ctx context.Context) (*models.EmailSMTPSettings, error)
	UpdateSMTPSettings(ctx context.Context, s *models.EmailSMTPSettings) error

	// Email Groups
	ListGroups(ctx context.Context) ([]models.EmailGroup, error)
	CreateGroup(ctx context.Context, g *models.EmailGroup, eventTypes []string) (int64, error)
	UpdateGroup(ctx context.Context, g *models.EmailGroup, eventTypes []string) error
	DeleteGroup(ctx context.Context, id int64) error
	GetGroupWithEvents(ctx context.Context, id int64) (*models.EmailGroup, []string, error)
	ListGroupsWithEvents(ctx context.Context) ([]EmailGroupWithEvents, error)

	// Sending
	SendTestEmail(ctx context.Context, to string) error
}

func NewEmailService(repo repository.EmailRepository) EmailService {
	return &emailServiceBase{repo: repo}
}

// ──────────────────────────────────────────────
// SMTP Settings
// ──────────────────────────────────────────────

func (s *emailServiceBase) GetSMTPSettings(ctx context.Context) (*models.EmailSMTPSettings, error) {
	return s.repo.GetSMTPSettings(ctx)
}

// UpdateSMTPSettings updates the SMTP configuration.
// An empty Password field means "keep the existing password".
func (s *emailServiceBase) UpdateSMTPSettings(ctx context.Context, settings *models.EmailSMTPSettings) error {
	if settings.Password == "" {
		existing, err := s.repo.GetSMTPSettings(ctx)
		if err == nil && existing != nil {
			settings.Password = existing.Password
		}
	}
	return s.repo.UpsertSMTPSettings(ctx, settings)
}

// ──────────────────────────────────────────────
// Email Groups
// ──────────────────────────────────────────────

func (s *emailServiceBase) ListGroups(ctx context.Context) ([]models.EmailGroup, error) {
	return s.repo.ListGroups(ctx)
}

func (s *emailServiceBase) CreateGroup(ctx context.Context, g *models.EmailGroup, eventTypes []string) (int64, error) {
	id, err := s.repo.CreateGroup(ctx, g)
	if err != nil {
		return 0, err
	}

	if len(eventTypes) > 0 {
		if err := s.repo.SetGroupEventTypes(ctx, id, eventTypes); err != nil {
			return id, err
		}
	}
	return id, nil
}

func (s *emailServiceBase) UpdateGroup(ctx context.Context, g *models.EmailGroup, eventTypes []string) error {
	if err := s.repo.UpdateGroup(ctx, g); err != nil {
		return err
	}
	return s.repo.SetGroupEventTypes(ctx, g.ID, eventTypes)
}

func (s *emailServiceBase) DeleteGroup(ctx context.Context, id int64) error {
	return s.repo.DeleteGroup(ctx, id)
}

func (s *emailServiceBase) GetGroupWithEvents(ctx context.Context, id int64) (*models.EmailGroup, []string, error) {
	g, err := s.repo.GetGroup(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	eventTypes, err := s.repo.GetGroupEventTypes(ctx, id)
	if err != nil {
		return g, nil, err
	}
	return g, eventTypes, nil
}

func (s *emailServiceBase) ListGroupsWithEvents(ctx context.Context) ([]EmailGroupWithEvents, error) {
	groups, err := s.repo.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]EmailGroupWithEvents, len(groups))
	for i, g := range groups {
		eventTypes, err := s.repo.GetGroupEventTypes(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		if eventTypes == nil {
			eventTypes = []string{}
		}
		result[i] = EmailGroupWithEvents{
			Group:      g,
			EventTypes: eventTypes,
		}
	}
	return result, nil
}

// ──────────────────────────────────────────────
// Sending
// ──────────────────────────────────────────────

// SendTestEmail builds a Sender from current DB settings and fires a test email.
func (s *emailServiceBase) SendTestEmail(ctx context.Context, to string) error {
	settings, err := s.repo.GetSMTPSettings(ctx)
	if err != nil {
		return err
	}

	sender := &email.Sender{
		Host:        settings.Host,
		Port:        settings.Port,
		Username:    settings.Username,
		Password:    settings.Password,
		SenderEmail: settings.SenderEmail,
		SenderName:  settings.SenderName,
		UseTLS:      settings.UseTLS,
	}

	return sender.SendTest(to)
}
