package ingest

import (
	"context"
	"database/sql"
	"log"
	"time"

	"nvr_core/db/models"
)

type BatchIngester struct {
	db         *sql.DB
	bufferChan chan *models.Segment
	batchSize  int
}

func NewBatchIngester(db *sql.DB, bufferSize int, batchSize int) *BatchIngester {
	return &BatchIngester{
		db:         db,
		bufferChan: make(chan *models.Segment, bufferSize),
		batchSize:  batchSize,
	}
}

// Enqueue is called by the IPC listener when a C++ worker finishes a 1-minute segment.
func (b *BatchIngester) Enqueue(seg *models.Segment) {
	b.bufferChan <- seg
}

// Start runs in the background, flushing the buffer to SQLite in transactions.
func (b *BatchIngester) Start(ctx context.Context) {
	var batch []*models.Segment
	ticker := time.NewTicker(3 * time.Second) // Flush at least every 3 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.flush(batch) // Flush remaining on shutdown
			return
		case seg := <-b.bufferChan:
			batch = append(batch, seg)
			if len(batch) >= b.batchSize {
				b.flush(batch)
				batch = nil // Reset batch
			}
		case <-ticker.C:
			if len(batch) > 0 {
				b.flush(batch)
				batch = nil
			}
		}
	}
}

func (b *BatchIngester) flush(batch []*models.Segment) {
	if len(batch) == 0 {
		return
	}

	tx, err := b.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	stmt, err := tx.Prepare(`INSERT INTO segments (camera_id, profile, start_time, end_time, file_path, size_bytes) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to prepare statement: %v", err)
		return
	}
	defer stmt.Close()

	for _, seg := range batch {
		if _, err := stmt.Exec(seg.CameraID, seg.Profile, seg.StartTime, seg.EndTime, seg.FilePath, seg.SizeBytes); err != nil {
			log.Printf("Failed to insert segment %s: %v", seg.FilePath, err)
			// Decide whether to rollback entirely or continue. For NVRs, continue is usually better.
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit batch: %v", err)
	}
}