package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"nvr_core/db/models"
)

// BookmarkRepository defines the contract for bookmark data access.
type BookmarkRepository interface {
	Create(ctx context.Context, bookmark *models.Bookmark) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Bookmark, error)
	GetByTimeRange(ctx context.Context, startUnix, endUnix int64) ([]*models.Bookmark, error)
	GetByCamera(ctx context.Context, cameraID int64, startUnix, endUnix int64) ([]*models.Bookmark, error)
	Update(ctx context.Context, bookmark *models.Bookmark) error
	Delete(ctx context.Context, id int64) error
}

type bookmarkRepo struct {
	db *sql.DB
}

// NewBookmarkRepository instantiates a new BookmarkRepository.
func NewBookmarkRepository(db *sql.DB) BookmarkRepository {
	return &bookmarkRepo{
		db: db,
	}
}

// Create inserts a new bookmark into the database and returns the generated ID.
func (r *bookmarkRepo) Create(ctx context.Context, b *models.Bookmark) (int64, error) {
	query := `
		INSERT INTO bookmarks (camera_id, user_id, start_time, duration, message)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at;
	`

	err := r.db.QueryRowContext(ctx, query, b.CameraID, b.UserID, b.Time, b.Duration, b.Message).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert bookmark: %w", err)
	}

	return b.ID, nil
}

// GetByID retrieves a single bookmark by its primary key.
func (r *bookmarkRepo) GetByID(ctx context.Context, id int64) (*models.Bookmark, error) {
	query := `
		SELECT id, camera_id, user_id, start_time, duration, message, created_at
		FROM bookmarks
		WHERE id = ?;
	`

	b := &models.Bookmark{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID,
		&b.CameraID,
		&b.UserID,
		&b.Time,
		&b.Duration,
		&b.Message,
		&b.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Or return a custom ErrNotFound
		}
		return nil, fmt.Errorf("failed to get bookmark by id: %w", err)
	}

	return b, nil
}

func (e *bookmarkRepo) scanRows(rows *sql.Rows) ([]*models.Bookmark, error) {

	var bookmarks []*models.Bookmark
	for rows.Next() {
		b := &models.Bookmark{}
		if err := rows.Scan(
			&b.ID,
			&b.CameraID,
			&b.UserID,
			&b.Time,
			&b.Duration,
			&b.Message,
			&b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bookmark row: %w", err)
		}
		bookmarks = append(bookmarks, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error in bookmarks: %w", err)
	}

	return bookmarks, nil

}

func (r *bookmarkRepo) GetByTimeRange(ctx context.Context, startUnix, endUnix int64) ([]*models.Bookmark, error) {
	query := `
		SELECT id, camera_id, user_id, start_time, duration, message, created_at
		FROM bookmarks
		WHERE start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, startUnix, endUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks by camera: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetByCamera retrieves all bookmarks for a specific camera within a time range.
func (r *bookmarkRepo) GetByCamera(ctx context.Context, cameraID int64, startUnix, endUnix int64) ([]*models.Bookmark, error) {
	query := `
		SELECT id, camera_id, user_id, start_time, duration, message, created_at
		FROM bookmarks
		WHERE camera_id = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, cameraID, startUnix, endUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks by camera: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// Update modifies an existing bookmark's duration and message.
func (r *bookmarkRepo) Update(ctx context.Context, b *models.Bookmark) error {
	query := `
		UPDATE bookmarks
		SET duration = ?, message = ?
		WHERE id = ?;
	`

	res, err := r.db.ExecContext(ctx, query, b.Duration, b.Message, b.ID)
	if err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("bookmark not found or no changes made")
	}

	return nil
}

// Delete removes a bookmark from the database.
func (r *bookmarkRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM bookmarks WHERE id = ?;`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	return nil
}