package repository

import (
	"context"
	"errors"
	"database/sql"

	"nvr_core/db/models"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameTaken = errors.New("username already exists")

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateRole(ctx context.Context, id int64, roleID int64) error
	UpdatePassword(ctx context.Context, id int64, newHash string) error
	Deactivate(ctx context.Context, id int64) error
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

// Create inserts a new user into the database.
func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, password, role_id, is_active) 
		VALUES (?, ?, ?, 1)
	`

	result, err := r.db.ExecContext(ctx, query, user.Username, user.Password, user.RoleID)
	if err != nil {
		// Basic SQLite unique constraint check
		if isUniqueConstraintViolation(err) {
			return ErrUsernameTaken
		}
		return err
	}

	user.ID, err = result.LastInsertId()
	return err
}

// GetByID fetches a user by their primary key.
func (r *userRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, password, role_id, is_active, created_at FROM users WHERE id = ?`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Password, &u.RoleID, &u.IsActive, &u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetByUsername is the workhorse for your Login API.
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password, role_id, is_active, created_at FROM users WHERE username = ?`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Password, &u.RoleID, &u.IsActive, &u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// UpdateRole changes the user's RBAC role.
func (r *userRepo) UpdateRole(ctx context.Context, id int64, roleID int64) error {
	query := `UPDATE users SET role_id = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, roleID, id)
	return err
}

// Update user password
func (r *userRepo) UpdatePassword(ctx context.Context, id int64, newHash string) error {
	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, newHash, id)
	return err
}


// Deactivate performs a soft-delete, immediately stripping their login access without destroying audit history.
func (r *userRepo) Deactivate(ctx context.Context, id int64) error {
	query := `UPDATE users SET is_active = 0 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}