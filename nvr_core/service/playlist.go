package service

import (
	"context"
	"fmt"

	"nvr_core/db/repository"
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
		return "", ErrVideoNotFound // Reusing the error we defined in playback.go
	}

	// Build the M3U8 Header
	playlist := m3u8.NewM3U8Builder(camID, baseURL)
	playlist.Begin()

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

	if len(segments) == 0 {
		return "", ErrVideoNotFound // Reusing the error we defined in playback.go
	}

	// Build the M3U8 Header
	playlist := m3u8.NewM3U8Builder(camID, baseURL)
	playlist.Begin()

	playlist.XVOD()
	playlist.XMediaSequence()
	playlist.XSetTargetDurationFor(segments)

	// Append each segment
	for _, seg := range segments {
		playlist.FeedVODSegment(seg)
	}

	// Close the playlist
	playlist.XVODEnd()

	return playlist.String(), nil
}