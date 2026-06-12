package repository

import (
	"context"
	"database/sql"
	"errors"

	"nvr_core/db/models"
)

var ErrTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository interface {
	// Create stores a new hashed refresh token [cite: 267]
	Create(ctx context.Context, token *models.RefreshToken) error
	
	// GetByHash retrieves a token for the /refresh endpoint validation [cite: 268, 269]
	GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	
	// Revoke invalidates a single specific session [cite: 262]
	Revoke(ctx context.Context, id string) error
	RevokeHashed(ctx context.Context, hash string) error
	
	// RevokeAllForUser instantly locks a user out of all active sessions across all devices 
	RevokeAllForUser(ctx context.Context, userID int64) error
	
	// DeleteExpired is meant for a background cron job to purge old data and keep the table small 
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepo struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

// Create inserts a new hashed refresh token into the database.
func (r *refreshTokenRepo) Create(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, user_agent, client_ip, is_revoked, expires_at, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, 
		token.ID, 
		token.UserID, 
		token.TokenHash, 
		token.UserAgent, 
		token.ClientIP, 
		token.IsRevoked, 
		token.ExpiresAt, 
		token.CreatedAt,
	)
	
	return err
}

// GetByHash retrieves a token by its hash. 
func (r *refreshTokenRepo) GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, user_agent, client_ip, is_revoked, expires_at, created_at 
		FROM refresh_tokens 
		WHERE token_hash = ?
	`

	var t models.RefreshToken
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.UserAgent, &t.ClientIP, &t.IsRevoked, &t.ExpiresAt, &t.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	
	return &t, nil
}

// Revoke flags a specific token as revoked without deleting the audit history.
func (r *refreshTokenRepo) Revoke(ctx context.Context, id string) error {
	query := `UPDATE refresh_tokens SET is_revoked = 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *refreshTokenRepo) RevokeHashed(ctx context.Context, hash string) error {
	query := `UPDATE refresh_tokens SET is_revoked = 1 WHERE token_hash = ?`
	_, err := r.db.ExecContext(ctx, query, hash)
	return err
}

// RevokeAllForUser acts as the ultimate kill-switch for a disabled or compromised account.
func (r *refreshTokenRepo) RevokeAllForUser(ctx context.Context, userID int64) error {
	query := `UPDATE refresh_tokens SET is_revoked = 1 WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired permanently deletes tokens that have passed their expiration date.
func (r *refreshTokenRepo) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.ExecContext(ctx, query)
	return err
}