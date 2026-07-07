package service

import (
	"context"
	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/utils"
)


type MaintainService interface {
	GetAbnormalSegments(ctx context.Context) ([]*models.Segment, error)
	FixSegmentsEndTime(ctx context.Context, segments []*models.Segment) int
}

func NewMaintainService(repo repository.SegmentRepository) MaintainService {
	return &maintainServiceBase{repo: repo}
}


func (ms *maintainServiceBase) GetAbnormalSegments(ctx context.Context) ([]*models.Segment, error) {
	return ms.repo.GetAbnormalDurationSegments(ctx)
}

func (ms *maintainServiceBase) FixSegmentsEndTime(ctx context.Context, segments []*models.Segment) int {

	ll := LOG.Prefix("[Maintain][FixSegmentsEndTime]")

	fixed := 0

	for _, seg := range segments {

		// Bypass snapshot, because there is no FilePath for it
		if seg.Profile == utils.SegmentSnapshotProfile {
			continue
		}

		actualDurationMs, err := utils.GetRealVideoDurationMs(seg.FilePath)
		if err != nil {
			ll.Info("Skipping segment", "segment", seg.ID, "error", err)
			continue
		}

		// Calculate the correct EndTime
		correctedEndTime := seg.StartTime + actualDurationMs

		// Update the database record
		err = ms.repo.UpdateSegmentEndTime(ctx, seg.ID, correctedEndTime)
		if err != nil {
			ll.Info("Failed to update DB for segment ", "segment", seg.ID, "error", err)
		} else {
			fixed++
			ll.Info("Fixed Segment",
				"segment", seg.ID,
				"old_duration", (seg.EndTime - seg.StartTime),
				"new_duration", actualDurationMs,
			)
		}
	}

	return fixed
}

// Get abnormal snapshots fixing it from it's source
func (ms *maintainServiceBase) FixSnapshotSegmentsEndTime(ctx context.Context) int {

}