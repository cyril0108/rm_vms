package repository

import (
	"context"
	"database/sql"
	"errors"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
)

/**
 * Sement Basic Operations
 */

// SegmentRepository defines the contract for segment data access.
type SegmentRepository interface {

	Insert(ctx context.Context, seg *models.Segment) error

	// Maintenance
	PruneOldest(ctx context.Context, limit int) ([]string, error)
	IncrementalVacuum(ctx context.Context, pages int) error

	// Segment search
	GetLastSegment(ctx context.Context) (*models.Segment, error)
	GetSegmentsByRange(ctx context.Context, camID int64, start, end int64) ([]*models.Segment, error)
	GetProfileSegmentsByRange(ctx context.Context, camID int64, profile string, start, end int64) ([]*models.Segment, error)
	GetSegmentAtTime(ctx context.Context, camID int64, profile string, timestamp int64) (*models.Segment, error)

	// Calendar
	GetDailySummary(ctx context.Context, camID int64, profile string, startUnix, endUnix int64) ([]dto.DailySummary, error)

	// Bulk Insert
	BulkInsert(ctx context.Context, segments []*models.Segment) error

	// Camera Data Check
	HasSegments(ctx context.Context, camID int64) (bool, error)

	// DB/file sanity check
	GetAllFilePaths(ctx context.Context) (map[string]struct{}, error)
	DeleteByFilePaths(ctx context.Context, paths []string) error
}

type segmentRepo struct {
	db *sql.DB
}

func NewSegmentRepository(db *sql.DB) SegmentRepository {
	return &segmentRepo{db: db}
}

func (r *segmentRepo) Insert(ctx context.Context, seg *models.Segment) error {
	query := `INSERT INTO segments (camera_id, profile, start_time, end_time, file_path, snapshot_path, size_bytes) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, seg.CameraID, seg.Profile, seg.StartTime, seg.EndTime, seg.FilePath, seg.SnapshotPath, seg.SizeBytes)
	if err != nil {
		return err
	}
	seg.ID, _ = result.LastInsertId()
	return nil
}

// PruneOldest executes the O(1) eviction policy and returns the physical file paths to delete.
func (r *segmentRepo) PruneOldest(ctx context.Context, limit int) ([]string, error) {
	query := `
		DELETE FROM segments 
		WHERE id IN (
			SELECT id FROM segments ORDER BY start_time ASC LIMIT ?
		)
		RETURNING file_path;
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

// IncrementalVacuum reclaims disk space from deleted rows.
func (r *segmentRepo) IncrementalVacuum(ctx context.Context, pages int) error {
	query := `PRAGMA incremental_vacuum(?);`
	_, err := r.db.ExecContext(ctx, query, pages)
	return err
}

func (r *segmentRepo) GetLastSegment(ctx context.Context) (*models.Segment, error) {
	query := `
		SELECT id, camera_id, profile, start_time, end_time, file_path, snapshot_path, size_bytes 
		FROM segments 
		ORDER BY start_time DESC 
		LIMIT 1
	`

	var seg models.Segment
	err := r.db.QueryRowContext(ctx, query).Scan(
		&seg.ID,
		&seg.CameraID,
		&seg.Profile,
		&seg.StartTime,
		&seg.EndTime,
		&seg.FilePath,
		&seg.SnapshotPath,
		&seg.SizeBytes,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil gracefully if the database is completely empty
		}
		return nil, err
	}

	return &seg, nil
}

func (r *segmentRepo) HasSegments(ctx context.Context, camID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM segments WHERE camera_id = ?)`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, camID).Scan(&exists)
	if err != nil {
		return false, err
	}
	
	return exists, nil
}