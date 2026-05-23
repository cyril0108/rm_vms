package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

func InitiateDB(ctx context.Context, dsn string) (*sql.DB, error) {

	dbConn, err := NewConnection(dsn)
	if err != nil {
		// log.Fatalf("Failed to open SQLite database: %v", err)
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Ensure tables are created
	if err := RunMigrations(ctx, dbConn); err != nil {
		// log.Fatalf()
		return nil, fmt.Errorf("Failed to run DB migrations: %v", err)
	}

	return dbConn, nil

}

// NewConnection opens the DB and applies critical performance pragmas.
func NewConnection(dsn string) (*sql.DB, error) {
	// Add URI parameters for connection pooling behavior
	dsnWithParams := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=auto_vacuum(INCREMENTAL)", dsn)

	db, err := sql.Open("sqlite", dsnWithParams)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize connection pool for concurrent NVR operations
	db.SetMaxOpenConns(25) // Allow concurrent readers/writers
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	return db, nil
}

// RunMigrations creates the tables and indexes.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS cameras (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		ip_address TEXT NOT NULL,
		http_port INTEGER DEFAULT 80,
		type TEXT NOT NULL DEFAULT 'onvif',
		username TEXT,
		password_enc TEXT,
		stream_url TEXT NOT NULL,
		sub_stream_url TEXT,
		onvif_profile_token TEXT,
		supports_ptz INTEGER DEFAULT 0,
		retention_gb_limit INTEGER,
		is_active INTEGER DEFAULT 1,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS segments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		camera_id TEXT NOT NULL,
		start_time INTEGER NOT NULL,
		end_time INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		size_bytes INTEGER NOT NULL,
		FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_segments_timeline ON segments(camera_id, start_time, end_time);
	CREATE INDEX IF NOT EXISTS idx_segments_pruning ON segments(start_time);
	`
	_, err := db.ExecContext(ctx, schema)
	return err
}