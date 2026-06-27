package service

import (
	"context"
	"os"
	"time"

    "nvr_core/db/repository"
    "nvr_core/hardware"
)

type RetentionService struct {
	path string
	repo repository.SegmentRepository
}

// StartDiskWatchdog should run as a goroutine from your main.go
func (s *RetentionService) StartDiskWatchdog(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.enforceDiskWatermark(ctx)
		}
	}
}

func (s *RetentionService) checkDiskUsage() (bool, error) {

	usage, err := hardware.GetDiskUsage(s.path)

	LOG.Info("[checkDiskUsage]", "usage", usage)

	if err != nil {
		return false, err
	}

	perc := (float64(usage.UsedBytes) / float64(usage.TotalBytes)) * 100

	LOG.Info("[checkDiskUsage]", "precentage", perc)

	if perc > 85 {
		return true, nil
	}

	return false, nil

}

func (s *RetentionService) enforceDiskWatermark(ctx context.Context) {
	// Check your physical disk usage here (e.g. using syscall.Statfs)
	isDiskFull, cdErr := s.checkDiskUsage() 

	if !isDiskFull || cdErr != nil {
		return // Disk is healthy or there is an error reading it, do nothing
	}

	LOG.Warn("Disk high watermark reached, beginning eviction")

	// The Pruning Loop
	for isDiskFull {
		// Evict in batches of 100 to prevent long database locks
		paths, err := s.repo.PruneOldest(ctx, 1)
		if err != nil {
			LOG.Error("Failed to prune database records", "error", err)
			return
		}

		if len(paths) == 0 {
			LOG.Info("Nothing to delete")
			break // Nothing left to delete
		}

		LOG.Info("Deleting ", "number", len(paths))

		// Delete the physical files
		for _, path := range paths {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				LOG.Error("Failed to delete physical video file", "path", path, "error", err)
			}
		}

		// Re-evaluate disk space before next iteration
		isDiskFull, cdErr = s.checkDiskUsage()
	}

	if cdErr != nil {
		LOG.Error("Last disk check failed", "error", cdErr)
	}

	// The Reclaim Step
	// Now that pruning is complete, release the SQLite pages back to the OS
	if err := s.repo.IncrementalVacuum(ctx, 100); err != nil {
		LOG.Error("Failed to incrementally vacuum database", "error", err)
	}

	LOG.Info("Disk space recovered successfully")
}