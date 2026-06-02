package service

import (
	"context"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
	"nvr_core/db/repository"
)

type TimelineService interface {
	GetContiguousBlocks(ctx context.Context, camID int64, start, end int64) ([]dto.TimelineBlock, error)
	GetProfileContiguousBlocks(ctx context.Context, camID int64, profile string, start, end int64) ([]dto.TimelineBlock, error)
	GetDailySummary(ctx context.Context, camID int64, profile string, start, end int64) ([]dto.DailySummary, error)
}

func NewTimelineService(repo repository.SegmentRepository) TimelineService {
	return &segmentServiceBase{repo: repo}
}

func (s *segmentServiceBase) GetProfileContiguousBlocks(ctx context.Context, camID int64, profile string, start, end int64) ([]dto.TimelineBlock, error) {
	segments, err := s.repo.GetProfileSegmentsByRange(ctx, camID, profile, start, end)
	LOG.Info("[GetProfileContiguousBlocks]", "segments", len(segments))
	if err != nil {
		return nil, err
	}
	return segmentsToTimeline(segments)
}

func (s *segmentServiceBase) GetContiguousBlocks(ctx context.Context, camID int64, start, end int64) ([]dto.TimelineBlock, error) {
	segments, err := s.repo.GetSegmentsByRange(ctx, camID, start, end)
	if err != nil {
		return nil, err
	}

	return segmentsToTimeline(segments)

}

func segmentsToTimeline(segments []*models.Segment) ([]dto.TimelineBlock, error) {

	LOG.Info("[segmentsToTimeline]", "segments", len(segments))

	if len(segments) == 0 {
		return []dto.TimelineBlock{}, nil
	}

	var blocks []dto.TimelineBlock

	// Start the first block
	currentBlock := dto.TimelineBlock{
		StartTime: segments[0].StartTime,
		EndTime:   segments[0].EndTime,
	}

	// FFmpeg segments aren't always exactly 60 seconds due to keyframe alignment.
	// We allow a 5-second gap between segments before considering it a true "break" in the recording.
	const gapToleranceSeconds = 5000

	for i := 1; i < len(segments); i++ {
		seg := segments[i]

		// If this segment starts within the tolerance window of the current block's end...
		if seg.StartTime <= (currentBlock.EndTime + gapToleranceSeconds) {
			// Extend the current block
			if seg.EndTime > currentBlock.EndTime {
				currentBlock.EndTime = seg.EndTime
			}
		} else {
			// The gap is too large. Finalize the current block and start a new one.
			currentBlock.ConvertToSeconds()
			blocks = append(blocks, currentBlock)
			currentBlock = dto.TimelineBlock{
				StartTime: seg.StartTime,
				EndTime:   seg.EndTime,
			}
		}
	}

	currentBlock.ConvertToSeconds()

	// Append the final block
	blocks = append(blocks, currentBlock)

	return blocks, nil

}

func (s *segmentServiceBase) GetDailySummary(ctx context.Context, camID int64, profile string, start, end int64) ([]dto.DailySummary, error) {
	return s.repo.GetDailySummary(ctx, camID, profile, start, end)
}