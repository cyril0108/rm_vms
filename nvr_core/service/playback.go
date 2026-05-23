package service

import (
	"context"
	"errors"
	"os"

	"nvr_core/db/repository"
)

var ErrVideoNotFound = errors.New("video not found at requested time")
var ErrFileMissing = errors.New("video file is missing from disk")

type PlaybackService interface {
	GetVideoFilePath(ctx context.Context, camID int64, profile string, timestamp int64) (string, error)
}

func NewPlaybackService(repo repository.SegmentRepository) PlaybackService {
	return &segmentServiceBase{repo: repo}
}

func (s *segmentServiceBase) GetVideoFilePath(ctx context.Context, camID int64, profile string, timestamp int64) (string, error) {

	// Ask the database which file contains this timestamp
	seg, err := s.repo.GetSegmentAtTime(ctx, camID, profile, timestamp)
	if err != nil {
		return "", err
	}
	if seg == nil {
		return "", ErrVideoNotFound
	}

	// Verify the file actually exists on the Linux filesystem (preventing 500 crashes)
	if _, err := os.Stat(seg.FilePath); os.IsNotExist(err) {
		// This happens if an admin manually deleted the file but the Reconciliation Loop hasn't run yet
		return "", ErrFileMissing
	}

	// Return ONLY the string path. The service doesn't stream bytes; the HTTP transport layer does.
	return seg.FilePath, nil
}