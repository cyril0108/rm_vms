package repository

import (
	"context"
	"fmt"
	"strings"

)

/**
 * DB/file sanity check
 */

// GetAllFilePaths returns a Hash Set of all file paths currently tracked in the DB.
func (r *segmentRepo) GetAllFilePaths(ctx context.Context) (map[string]struct{}, error) {
	query := `SELECT file_path FROM segments`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query file paths: %w", err)
	}
	defer rows.Close()

	// Using an empty struct{} consumes 0 bytes of memory in Go
	paths := make(map[string]struct{})
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths[path] = struct{}{}
	}
	return paths, rows.Err()
}

func (r *segmentRepo) GetAllSnapshotPaths(ctx context.Context) (map[string]struct{}, error) {
	query := `SELECT snapshot_path FROM segments`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query file paths: %w", err)
	}
	defer rows.Close()

	// Using an empty struct{} consumes 0 bytes of memory in Go
	paths := make(map[string]struct{})
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths[path] = struct{}{}
	}
	return paths, rows.Err()
}

// DeleteByFilePaths removes ghost records in batches.
func (r *segmentRepo) DeleteByFilePaths(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// Batch delete in chunks of 900 to respect SQLite's variable limit
	chunkSize := 900
	for i := 0; i < len(paths); i += chunkSize {
		end := i + chunkSize
		if end > len(paths) {
			end = len(paths)
		}
		chunk := paths[i:end]

		placeholders := strings.Repeat("?,", len(chunk)-1) + "?"
		query := fmt.Sprintf(`DELETE FROM segments WHERE file_path IN (%s)`, placeholders)

		args := make([]interface{}, len(chunk))
		for j, p := range chunk {
			args[j] = p
		}

		if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("failed to delete ghost records: %w", err)
		}
	}
	return nil
}

func (r *segmentRepo) DeleteBySnapshotPaths(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// Batch delete in chunks of 900 to respect SQLite's variable limit
	chunkSize := 900
	for i := 0; i < len(paths); i += chunkSize {
		end := i + chunkSize
		if end > len(paths) {
			end = len(paths)
		}
		chunk := paths[i:end]

		placeholders := strings.Repeat("?,", len(chunk)-1) + "?"
		query := fmt.Sprintf(`DELETE FROM segments WHERE snapshot_path IN (%s)`, placeholders)

		args := make([]interface{}, len(chunk))
		for j, p := range chunk {
			args[j] = p
		}

		if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("failed to delete ghost records: %w", err)
		}
	}
	return nil
}