package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"nvr_core/db/models" // Adjust this import based on where your License struct lives
)

var ErrLicenseNotFound = errors.New("license not found")

// LicenseRepository defines the database operations for system licenses
type LicenseRepository interface {
	Create(ctx context.Context, lic *models.License) error
	GetAll(ctx context.Context) ([]*models.License, error)
	GetValidLicenses(ctx context.Context) ([]*models.License, error)
	Delete(ctx context.Context, id int64) error
}

type licenseRepo struct {
	db *sql.DB
}

// NewLicenseRepository instantiates a new license repository
func NewLicenseRepository(db *sql.DB) LicenseRepository {
	return &licenseRepo{db: db}
}

// Create inserts a newly uploaded, cryptographically verified license into the database.
func (r *licenseRepo) Create(ctx context.Context, lic *models.License) error {
	query := `
		INSERT INTO licenses (raw_token, iss, aud, kind, machine_id, max_devices, issued_at, expires_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		lic.RawToken,
		lic.Iss,
		lic.Aud,
		lic.Kind,
		lic.MachineID,
		lic.MaxDevices,
		lic.IssuedAt,
		lic.ExpiresAt,
	)
	if err != nil {
		return err
	}

	// Capture the generated ID and attach it back to the struct
	id, err := result.LastInsertId()
	if err == nil {
		lic.ID = id
	}

	return nil
}

// GetAll retrieves all licenses from the database.
// This is used by the Go Manager on startup to enforce rules, and by the UI to list active licenses.
func (r *licenseRepo) GetAll(ctx context.Context) ([]*models.License, error) {
	query := `
		SELECT id, raw_token, iss, aud, kind, machine_id, max_devices, issued_at, expires_at, uploaded_at
		FROM licenses
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.rowsToLicenses(rows)
}

func (r *licenseRepo) GetValidLicenses(ctx context.Context) ([]*models.License, error) {
	query := `
		SELECT id, raw_token, iss, aud, kind, machine_id, max_devices, issued_at, expires_at, uploaded_at
		FROM licenses
		WHERE expires_at > ?
		ORDER BY uploaded_at DESC
	`
	// Get the exact current Unix timestamp from the Go runtime
	now := time.Now().Unix()

	rows, err := r.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.rowsToLicenses(rows)
}


func (r *licenseRepo) rowsToLicenses(rows *sql.Rows) ([]*models.License, error) {

	var licenses []*models.License

	for rows.Next() {
		var lic models.License
		var uploadedAt string // SQLite stores dates as strings/text by default

		err := rows.Scan(
			&lic.ID,
			&lic.RawToken,
			&lic.Iss,
			&lic.Aud,
			&lic.Kind,
			&lic.MachineID,
			&lic.MaxDevices,
			&lic.IssuedAt,
			&lic.ExpiresAt,
			&uploadedAt,
		)
		if err != nil {
			return nil, err
		}

		parsedTime, err := time.Parse("2006-01-02T15:04:05Z", uploadedAt)
		if err != nil {
		    LOG.Error("Failed to parse UploadedAt time", "error", err, "input", uploadedAt)
		} else {
		    lic.UploadedAt = parsedTime
		}

		licenses = append(licenses, &lic)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return licenses, nil
}



// Delete removes a specific license by its ID.
func (r *licenseRepo) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM licenses 
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrLicenseNotFound
	}

	return nil
}