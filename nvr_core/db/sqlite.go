package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// InitiateDB opens the connection pool and safely applies pending migrations.
func InitiateDB(ctx context.Context, dsn string) (*sql.DB, error) {

	dbConn, err := NewConnection(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := RunMigrations(ctx, dbConn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	return dbConn, nil
}

// NewConnection opens the DB and applies critical performance pragmas.
func NewConnection(dsn string) (*sql.DB, error) {
	// Note: We include _pragma=foreign_keys(1) here to ensure cascading deletes work!
	dsnWithParams := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=auto_vacuum(INCREMENTAL)&_pragma=foreign_keys(1)", dsn)

	db, err := sql.Open("sqlite", dsnWithParams)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	return db, nil
}

// RunMigrations executes all unapplied .up.sql files from the embedded filesystem.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	// Prepare the embedded filesystem for the migration engine
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to load embedded migration files: %w", err)
	}

	// Prepare the SQLite database driver for the migration engine
	dbDriver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize sqlite migration driver: %w", err)
	}

	// Initialize the migrator
	migrator, err := migrate.NewWithInstance(
		"iofs", sourceDriver, 
		"sqlite", dbDriver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	// Execute all pending "Up" migrations
	err = migrator.Up()

	// 'ErrNoChange' just means the database is already fully up-to-date
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}