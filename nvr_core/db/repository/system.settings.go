package repository

import (
	"context"
	"database/sql"
	"errors"
)

var ErrSettingNotFound = errors.New("setting not found")

type SystemSettingsRepository interface {
	// Create a new setting
	Create(ctx context.Context, key string, value string) error

	// GetByHash retrieves a token for the /refresh endpoint validation [cite: 268, 269]
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)

}

type systemSettingsRepo struct {
	db *sql.DB
}

func NewSystemSettingsRepository(db *sql.DB) SystemSettingsRepository {
	return &systemSettingsRepo{db: db}
}

// Create inserts a new hashed refresh token into the database.
func (r *systemSettingsRepo) Create(ctx context.Context, key string, value string) error {
	query := `
		INSERT INTO system_settings (key, value) 
		VALUES (?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, 
		key, 
		value,
	)
	
	return err
}

// GetByHash retrieves a token by its hash. 
func (r *systemSettingsRepo) Get(ctx context.Context, key string) (string, error) {
	query := `
		SELECT value
		FROM system_settings 
		WHERE key = ?
	`

	var val string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&val)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrTokenNotFound
		}
		return "", err
	}

	return val, nil
}

// 
func (r *systemSettingsRepo) Set(ctx context.Context, key string, value string) error {
	query := `UPDATE system_settings SET value = ? WHERE key = ?`
	_, err := r.db.ExecContext(ctx, query, value, key)
	return err
}

