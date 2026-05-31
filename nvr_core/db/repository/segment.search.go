package repository

import (
	"context"
	"database/sql"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
)

func (r *segmentRepo) GetProfileSegmentsByRange(ctx context.Context, camID int64, profile string, start, end int64) ([]*models.Segment, error) {

	log := LOG.Lin("camID", camID, "profile", profile, "start", start, "end", end)

	log.Info("[GetProfileSegmentsByRange]")

	query := `
		SELECT id, camera_id, profile, start_time, end_time, file_path, snapshot_path, size_bytes 
		FROM segments 
		WHERE camera_id = ? AND profile = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, camID, profile, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []*models.Segment
	for rows.Next() {
		var seg models.Segment
		if err := rows.Scan(&seg.ID, &seg.CameraID, &seg.Profile, &seg.StartTime, &seg.EndTime, &seg.FilePath, &seg.SnapshotPath, &seg.SizeBytes); err != nil {
			return nil, err
		}
		segments = append(segments, &seg)
	}

	log.Info("[GetProfileSegmentsByRange]", "segments", len(segments))

	return segments, rows.Err()
}


func (r *segmentRepo) GetSegmentsByRange(ctx context.Context, camID int64, start, end int64) ([]*models.Segment, error) {
	query := `
		SELECT id, camera_id, profile, start_time, end_time, file_path, size_bytes 
		FROM segments 
		WHERE camera_id = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query, camID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []*models.Segment
	for rows.Next() {
		var seg models.Segment
		if err := rows.Scan(&seg.ID, &seg.CameraID, &seg.Profile, &seg.StartTime, &seg.EndTime, &seg.FilePath, &seg.SizeBytes); err != nil {
			return nil, err
		}
		segments = append(segments, &seg)
	}
	return segments, rows.Err()
}

func (r *segmentRepo) GetSegmentAtTime(ctx context.Context, camID int64, profile string, timestamp int64) (*models.Segment, error) {
	// Find the segment where the requested timestamp falls between start and end.
	query := `
		SELECT id, camera_id, profile, start_time, end_time, file_path, snapshot_path, size_bytes 
		FROM segments 
		WHERE camera_id = ? AND profile = ? AND start_time <= ? AND end_time >= ?
		LIMIT 1
	`

	var seg models.Segment
	err := r.db.QueryRowContext(ctx, query, camID, profile, timestamp, timestamp).Scan(
		&seg.ID, &seg.CameraID, &seg.Profile, &seg.StartTime, &seg.EndTime, &seg.FilePath, &seg.SnapshotPath, &seg.SizeBytes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No video exists at this exact second
		}
		return nil, err
	}

	return &seg, nil
}

// GetDailySummary(ctx context.Context, camID string, startUnix, endUnix int64) ([]dto.DailySummary, error)
func (r *segmentRepo) GetDailySummary(ctx context.Context, camID int64, profile string, startUnix, endUnix int64) ([]dto.DailySummary, error) {
	// The 'localtime' modifier tells SQLite to convert the Unix epoch into the 
	// server's local timezone before extracting the YYYY-MM-DD date.
	query := `
		SELECT 
			strftime('%Y-%m-%d', start_time / 1000, 'unixepoch', 'localtime') AS record_date,
			SUM(end_time - start_time) / 1000 AS total_seconds
		FROM segments
		WHERE camera_id = ? AND profile = ? AND start_time >= ? AND end_time <= ?
		GROUP BY record_date
		ORDER BY record_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, camID, profile, startUnix, endUnix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []dto.DailySummary
	for rows.Next() {
		var s dto.DailySummary
		if err := rows.Scan(&s.Date, &s.TotalSeconds); err != nil {
			return nil, err
		}

		// Format the seconds into HH:MM:SS in Go
		// h := s.TotalSeconds / 3600
		// m := (s.TotalSeconds % 3600) / 60
		// sec := s.TotalSeconds % 60
		// s.Formatted = fmt.Sprintf("%02d:%02d:%02d", h, m, sec)

		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}