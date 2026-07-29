package service

import (
	"context"
	"os"
	"time"

	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/events"
	"nvr_core/hardware"
	"nvr_core/i18n/tw"
	"nvr_core/utils"
)

var Lang = tw.Translator{}

type ActiveCameraProvider func() int

type RetentionService struct {
	path string
	repo repository.SegmentRepository
	getActiveCameras ActiveCameraProvider
	eventHub   *events.EventHub
}

const HighWaterMark = utils.HighWaterMark
const LowWaterMark = utils.LowWaterMark

// cameraProvider ActiveCameraProvider

func NewRetentionService(segRepo repository.SegmentRepository, path string, evHub *events.EventHub, cameraProvider ActiveCameraProvider) RetentionService {
	return RetentionService{
		path: path,
		repo: segRepo,
		getActiveCameras: cameraProvider,
		eventHub: evHub,
	}
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

func (s *RetentionService) checkDiskUsage() (float64, error) {

	usage, err := hardware.GetDiskUsage(s.path)

	if err != nil {
		return 0, err
	}

	perc := float64(usage.UsedBytes) / float64(usage.TotalBytes)

	return perc, nil

}

func (s *RetentionService) SendDiskNearFullWarning(per float64) {
	msg := Lang.Translate(tw.MSGDiskNearFullWarning, LowWaterMark, per)
	(*s.eventHub).SendEvent(models.EventTypeDiskWarning, msg, nil)
}

func (s *RetentionService) SendDiskFullWarning() {
	msg := Lang.Translate(tw.MSGDiskFullWarning)
	(*s.eventHub).SendEvent(models.EventTypeDiskWarning, msg, nil)
}

func (s *RetentionService) enforceDiskWatermark(ctx context.Context) {

	// Check your physical disk usage here (e.g. using syscall.Statfs)
	diskPerc, cdErr := s.checkDiskUsage() 

	if (diskPerc < HighWaterMark) || cdErr != nil {

		if diskPerc > LowWaterMark {
			s.SendDiskNearFullWarning(diskPerc)
		}

		return // Disk is healthy or there is an error reading it, do nothing
	}

	go s.SendDiskFullWarning()

	// One cameras could have 3 segments (main/sub/snapshot)
	batchN := s.getActiveCameras() * 3

	LOG.Warn("[enforceDiskWatermark] Disk high watermark reached, beginning eviction", 
		"precentage", diskPerc,
		"batchNumber", batchN)

	// The Pruning Loop
	for (diskPerc > LowWaterMark) {
		// Evict in batches of 100 to prevent long database locks
		paths, err := s.repo.PruneOldest(ctx, batchN)
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
		diskPerc, cdErr = s.checkDiskUsage()
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