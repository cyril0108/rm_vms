package service

import (
	"context"
	"fmt"

	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/utils"
	"nvr_core/utils/m3u8"
)

type PlaylistService interface {
	// GeneratePlaylist creates an M3U8 playlist string for a specific time range.
	// baseURL is injected by the HTTP handler so the service doesn't need to know the server's IP.
	GeneratePlaylist(ctx context.Context, camID int64, profile string, start, end int64, baseURL string) (string, error)
	// GeneratePlaylist creates an M3U8 VOD playlist string for a specific time range.
	GenerateVODPlaylist(ctx context.Context, camID int64, profile string, start, end int64, baseURL string) (string, error)
}

func NewPlaylistService(repo repository.SegmentRepository) PlaylistService {
	return &segmentServiceBase{repo: repo}
}

func (s *segmentServiceBase) GeneratePlaylist(ctx context.Context, camID int64, profile string, start, end int64, baseURL string) (string, error) {
	// Fetch all segments within the requested time window
	segments, err := s.repo.GetProfileSegmentsByRange(ctx, camID, profile, start, end)
	if err != nil {
		return "", fmt.Errorf("failed to fetch segments for playlist: %w", err)
	}

	if len(segments) == 0 {
		return "", ErrVideoSegmentNotFound // Reusing the error we defined in playback.go
	}

	// Build the M3U8 Header
	playlist := m3u8.NewM3U8Builder(camID, baseURL)
	playlist.Begin()
	playlist.XSetTargetDurationFor(segments)

	// Append each segment
	for _, seg := range segments {
		playlist.FeedSegment(seg)
	}

	return playlist.String(), nil
}

func (s *segmentServiceBase) GenerateVODPlaylist(ctx context.Context, camID int64, profile string, start, end int64, baseURL string) (string, error) {
	// Fetch all segments within the requested time window
	segments, err := s.repo.GetProfileSegmentsByRange(ctx, camID, profile, start, end)
	if err != nil {
		return "", fmt.Errorf("failed to fetch segments for playlist: %w", err)
	}

	limit := len(segments)
	if limit == 0 {
		return "", ErrVideoSegmentNotFound // Reusing the error we defined in playback.go
	}

	fst := segments[0]
	fst.StartTime = start

	lstIdx := limit-1
	lst := segments[lstIdx]
	if lst.EndTime > end {
		lst.EndTime = end
	}

	// Get resolution from first segment file
	referenceFilePath := fst.FilePath
	resolution, err := utils.GetVideoResolution(referenceFilePath)
	if err != nil {
	    // Fallback to a default if probe fails, or the profile's expected default
	    resolution = "1280x720" 
	}

	// Build the M3U8 Header
	playlist := m3u8.NewM3U8Builder(camID, baseURL)
	playlist.Begin()

	playlist.XVOD()
	playlist.XMediaSequence()
	playlist.XSetTargetDurationFor(segments)

	var lastSeg *models.Segment

	// Threshold, gap lower than gapTolerance will be seen as
	// continuing records
	gapTolerance := int64(300)

	// Append each segment
	for i, seg := range segments {

		if lastSeg != nil {
			if gap := seg.StartTime - lastSeg.EndTime; gap > gapTolerance {
				playlist.FeedVODGap(seg.CameraID, lastSeg.EndTime, gap, resolution)
			}
		}

		if i != lstIdx {

			playlist.FeedVODSegment(seg)

		} else {

			playlist.FeedVODSegmentDuration(seg)

		}
		lastSeg = seg
	}

	// Close the playlist
	playlist.XVODEnd()

	return playlist.String(), nil
}