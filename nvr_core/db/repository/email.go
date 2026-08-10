package repository

import (
	"context"
	"database/sql"

	"nvr_core/db/models"
)

type EmailRepository interface {
	// SMTP Settings (singleton row)
	GetSMTPSettings(ctx context.Context) (*models.EmailSMTPSettings, error)
	UpsertSMTPSettings(ctx context.Context, s *models.EmailSMTPSettings) error

	// Email Groups
	ListGroups(ctx context.Context) ([]models.EmailGroup, error)
	GetGroup(ctx context.Context, id int64) (*models.EmailGroup, error)
	CreateGroup(ctx context.Context, g *models.EmailGroup) (int64, error)
	UpdateGroup(ctx context.Context, g *models.EmailGroup) error
	DeleteGroup(ctx context.Context, id int64) error

	// Group ↔ Event Type mappings
	GetGroupEventTypes(ctx context.Context, groupID int64) ([]string, error)
	SetGroupEventTypes(ctx context.Context, groupID int64, eventTypes []string) error
	GetGroupsByEventType(ctx context.Context, eventType string) ([]models.EmailGroup, error)
}

type emailRepo struct {
	db *sql.DB
}

func NewEmailRepository(db *sql.DB) EmailRepository {
	return &emailRepo{db: db}
}

// ──────────────────────────────────────────────
// SMTP Settings
// ──────────────────────────────────────────────

func (r *emailRepo) GetSMTPSettings(ctx context.Context) (*models.EmailSMTPSettings, error) {
	query := `
		SELECT id, host, port, username, password, sender_email, sender_name,
		       use_tls, enabled, updated_at
		FROM email_smtp_settings
		WHERE id = 1
	`

	var s models.EmailSMTPSettings
	err := r.db.QueryRowContext(ctx, query).Scan(
		&s.ID, &s.Host, &s.Port, &s.Username, &s.Password,
		&s.SenderEmail, &s.SenderName, &s.UseTLS, &s.Enabled, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *emailRepo) UpsertSMTPSettings(ctx context.Context, s *models.EmailSMTPSettings) error {
	query := `
		UPDATE email_smtp_settings
		SET host = ?, port = ?, username = ?, password = ?,
		    sender_email = ?, sender_name = ?, use_tls = ?, enabled = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`
	_, err := r.db.ExecContext(ctx, query,
		s.Host, s.Port, s.Username, s.Password,
		s.SenderEmail, s.SenderName, s.UseTLS, s.Enabled,
	)
	return err
}

// ──────────────────────────────────────────────
// Email Groups
// ──────────────────────────────────────────────

func (r *emailRepo) ListGroups(ctx context.Context) ([]models.EmailGroup, error) {
	query := `SELECT id, name, recipients, created_at, updated_at FROM email_groups ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.EmailGroup
	for rows.Next() {
		var g models.EmailGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Recipients, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *emailRepo) GetGroup(ctx context.Context, id int64) (*models.EmailGroup, error) {
	query := `SELECT id, name, recipients, created_at, updated_at FROM email_groups WHERE id = ?`

	var g models.EmailGroup
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&g.ID, &g.Name, &g.Recipients, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *emailRepo) CreateGroup(ctx context.Context, g *models.EmailGroup) (int64, error) {
	query := `INSERT INTO email_groups (name, recipients) VALUES (?, ?)`

	result, err := r.db.ExecContext(ctx, query, g.Name, g.Recipients)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *emailRepo) UpdateGroup(ctx context.Context, g *models.EmailGroup) error {
	query := `UPDATE email_groups SET name = ?, recipients = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, g.Name, g.Recipients, g.ID)
	return err
}

func (r *emailRepo) DeleteGroup(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM email_groups WHERE id = ?`, id)
	return err
}

// ──────────────────────────────────────────────
// Group ↔ Event Type mappings
// ──────────────────────────────────────────────

func (r *emailRepo) GetGroupEventTypes(ctx context.Context, groupID int64) ([]string, error) {
	query := `SELECT event_type FROM email_group_events WHERE group_id = ?`

	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

// SetGroupEventTypes replaces all event type mappings for a group in a single transaction.
func (r *emailRepo) SetGroupEventTypes(ctx context.Context, groupID int64, eventTypes []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing mappings
	if _, err := tx.ExecContext(ctx, `DELETE FROM email_group_events WHERE group_id = ?`, groupID); err != nil {
		return err
	}

	// Insert new mappings
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO email_group_events (group_id, event_type) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, et := range eventTypes {
		if _, err := stmt.ExecContext(ctx, groupID, et); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetGroupsByEventType finds all groups that are subscribed to a given event type.
// Used by the EmailNotifier to determine who should receive a notification.
func (r *emailRepo) GetGroupsByEventType(ctx context.Context, eventType string) ([]models.EmailGroup, error) {
	query := `
		SELECT g.id, g.name, g.recipients, g.created_at, g.updated_at
		FROM email_groups g
		INNER JOIN email_group_events ge ON g.id = ge.group_id
		WHERE ge.event_type = ?
	`

	rows, err := r.db.QueryContext(ctx, query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.EmailGroup
	for rows.Next() {
		var g models.EmailGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Recipients, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}
