package service

import (
	"context"
	"sort"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
	"nvr_core/db/repository"
)

type TimelineService interface {
	GetProfileSegmentItems(ctx context.Context, camID int64, profile string, start, end int64) ([]*dto.SegmentItem, error)
	GetCameraSnapshots(ctx context.Context, camID int64, start, end int64) ([]*dto.SegmentSnapshot, error)
	GetContiguousBlocks(ctx context.Context, camID int64, start, end int64) ([]dto.TimelineBlock, error)
	GetProfileContiguousBlocks(ctx context.Context, camID int64, profile string, start, end int64) ([]dto.TimelineBlock, error)
	GetDailySummary(ctx context.Context, camID int64, profile string, start, end int64) ([]dto.DailySummary, error)
}

func NewTimelineService(repo repository.SegmentRepository) TimelineService {
	return &segmentServiceBase{repo: repo}
}

func (s *segmentServiceBase) GetProfileSegmentItems(ctx context.Context, camID int64, profile string, start, end int64) ([]*dto.SegmentItem, error) {
	segments, err := s.repo.GetProfileSegmentsByRange(ctx, camID, profile, start, end)
	LOG.Info("[GetProfileSegmentItems]", "segments", len(segments))
	if err != nil {
		return nil, err
	}
	return segmentsToDTOItems(segments)
}

func (s *segmentServiceBase) GetCameraSnapshots(ctx context.Context, camID int64, start, end int64) ([]*dto.SegmentSnapshot, error) {
	segments, err := s.repo.GetProfileSegmentsByRange(ctx, camID, models.SegmentSnapshotProfile, start, end)
	LOG.Info("[GetProfileSegmentItems]", "segments", len(segments))
	if err != nil {
		return nil, err
	}
	return segmentsToSnapshots(segments)
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

func (s *segmentServiceBase) GetDailySummary(ctx context.Context, camID int64, profile string, start, end int64) ([]dto.DailySummary, error) {
	return s.repo.GetDailySummary(ctx, camID, profile, start, end)
}

// ---------------
// Private methods
// ---------------

func segmentsToDTOItems(segments []*models.Segment) ([]*dto.SegmentItem, error) {

	ll := LOG.Prefix("[segmentsToDTOItems]")

	ll.Info("segments count", "n", len(segments))

	if len(segments) == 0 {
		return []*dto.SegmentItem{}, nil
	}

	var list []*dto.SegmentItem

	for _, seg := range segments {
		it := dto.NewSegmentItemFrom(seg)
		it.ConvertToSeconds()
		list = append(list, it)
	}

	return list, nil

}


func segmentsToSnapshots(segments []*models.Segment) ([]*dto.SegmentSnapshot, error) {

	ll := LOG.Prefix("[segmentsToSnapshots]")

	ll.Info("segments count", "n", len(segments))

	if len(segments) == 0 {
		return []*dto.SegmentSnapshot{}, nil
	}

	var list []*dto.SegmentSnapshot

	for _, seg := range segments {
		list = append(list, dto.NewSegmentSnapshotFrom(seg))
	}

	return list, nil

}

func segmentsToTimeline(segments []*models.Segment) ([]dto.TimelineBlock, error) {

	limit := len(segments)
	LOG.Info("[segmentsToTimeline]", "segments", limit)

	if limit == 0 {
		return []dto.TimelineBlock{}, nil
	}

	// This ensures that interleaved segments from multiple profiles 
	// are processed in chronological order so overlaps merge correctly.
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].StartTime < segments[j].StartTime
	})

	var blocks []dto.TimelineBlock

	// Start the first block
	currentBlock := dto.TimelineBlock{
		TimeRange: dto.TimeRange{
			StartTime: segments[0].StartTime,
			EndTime:   segments[0].EndTime,
		},
	}

	// FFmpeg segments aren't always exactly 60 seconds due to keyframe alignment.
	// We allow a 5-second gap between segments before considering it a true "break" in the recording.
	const gapToleranceSeconds = 5000

	for i := 1; i < limit; i++ {
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
				TimeRange: dto.TimeRange{
					StartTime: seg.StartTime,
					EndTime:   seg.EndTime,
				},
			}
		}
	}

	currentBlock.ConvertToSeconds()

	// Append the final block
	blocks = append(blocks, currentBlock)

	return blocks, nil

}

