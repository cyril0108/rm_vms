package repository

import (
	"context"
	"fmt"
	"nvr_core/db/models"
)

func (r *segmentRepo) BulkInsert(ctx context.Context, segments []*models.Segment) error {
	if len(segments) == 0 {
		return nil
	}

	// Begin Transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Prepare Statement
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO segments (camera_id, profile, start_time, end_time, file_path, size_bytes) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute Batch
	for _, seg := range segments {
		if _, err := stmt.ExecContext(ctx, seg.CameraID, seg.Profile, seg.StartTime, seg.EndTime, seg.FilePath, seg.SizeBytes); err != nil {
			// Depending on strictness, you might choose to log and continue, but Rollback is safer for data integrity
			tx.Rollback()
			return fmt.Errorf("failed to insert segment %s: %w", seg.FilePath, err)
		}
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch: %w", err)
	}

	return nil
}