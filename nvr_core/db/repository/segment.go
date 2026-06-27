package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	GetHourlyBurnRateBytes(ctx context.Context) (uint64, error)

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


// ================================================
// Maintenance
// ================================================

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
	// query := `PRAGMA incremental_vacuum(?);`
	// PRAGMA statements do not support parameter binding. 
	// We safely format the integer directly into the query.
	query := fmt.Sprintf("PRAGMA incremental_vacuum(%d);", pages)
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// GetHourlyBurnRateBytes calculates the total video bytes written to disk in the last 3600 seconds.
// This is used alongside syscall.Statfs to calculate real-time remaining recording capacity.
func (r *segmentRepo) GetHourlyBurnRateBytes(ctx context.Context) (uint64, error) {
	// Calculate the threshold in Go (ensures exact clock sync with the rest of your app)
	oneHourAgo := time.Now().Unix() - 3600

	// The Query
	// CRITICAL: We wrap SUM() in COALESCE(). If the NVR was just turned on 
	// and there are zero segments in the last hour, SUM() returns NULL. 
	// Scanning NULL into a Go integer will cause a runtime panic. 
	// COALESCE forces it to safely return 0 instead.
	query := `
		SELECT COALESCE(SUM(size_bytes), 0)
		FROM segments
		WHERE start_time >= ?;
	`

	var burnRateBytes uint64

	// Execute the query with a context timeout
	err := r.db.QueryRowContext(ctx, query, oneHourAgo).Scan(&burnRateBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate hourly burn rate: %w", err)
	}

	return burnRateBytes, nil
}

