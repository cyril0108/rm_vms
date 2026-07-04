package service

import (
	"context"
	"fmt"
	"nvr_core/db/repository"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
	// "nvr_core/db/models"
)

type ExportService interface {
	ExportTimeRange(ctx context.Context, rootPath string, camID int64, profile string, reqStart, reqEnd int64) (string, error)
}

func NewExportService(repo repository.SegmentRepository) ExportService {
	return &segmentServiceBase{repo: repo}
}


func (s *segmentServiceBase) ExportTimeRange(ctx context.Context, rootPath string, camID int64, profile string, reqStart, reqEnd int64) (string, error) {

	// Fetch segments intersecting the requested time range
	segments, err := s.repo.GetProfileSegmentsByRange(ctx, camID, profile, reqStart, reqEnd)
	if err != nil {
		return "", err
	}
	if len(segments) == 0 {
		return "", fmt.Errorf("no recordings found in the specified time range")
	}

	// Sort segments chronologically
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].StartTime < segments[j].StartTime
	})

	// Prepare the FFmpeg concat text file
	concatFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("concat_%d.txt", reqStart))
	file, err := os.Create(concatFilePath)
	if err != nil {
		return "", err
	}
	defer os.Remove(concatFilePath) // Clean up instruction file after FFmpeg runs
	defer file.Close()

	// Generate the instructions
	for i, seg := range segments {
		// Write the file path (must be safely escaped if paths have spaces, but standard paths are fine)
		fmt.Fprintf(file, "file '%s'\n", seg.FilePath)

		// First segment: Trim the beginning if the user requested a start time halfway through
		if i == 0 && reqStart > seg.StartTime {
			offsetSeconds := float64(reqStart-seg.StartTime) / 1000.0
			fmt.Fprintf(file, "inpoint %.3f\n", offsetSeconds)
		}

		// Last segment: Trim the end if the user requested an end time halfway through
		if i == len(segments)-1 && reqEnd < seg.EndTime {
			durationSeconds := float64(reqEnd-seg.StartTime) / 1000.0
			fmt.Fprintf(file, "outpoint %.3f\n", durationSeconds)
		}
	}
	file.Sync()

	// Execute FFmpeg
	// Define the output file path (e.g., in a dedicated public/exports directory)
	exportDir := filepath.Join(rootPath, "export")
	// 0755: Owner can read/write/execute. Group/Others can read/execute.
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory at %s: %v", exportDir, err)
	}
	outputPath := filepath.Join(exportDir, fmt.Sprintf("export_cam%d_%d.mp4", camID, reqStart))

	// -f concat: Use the concat demuxer
	// -safe 0: Allow absolute file paths in the text file
	// -c copy: STREAM COPY (No CPU re-encoding)
	// -movflags +faststart: Moves MOOV atom to the front so the exported MP4 can stream instantly over HTTP
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y", // Overwrite output
		"-f", "concat",
		"-safe", "0",
		"-i", concatFilePath,
		"-c", "copy",
		"-movflags", "+faststart",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg export failed: %v", err)
	}

	return outputPath, nil
}


// StartExportCleanupWatchdog runs in the background and deletes old exports
func StartExportCleanupWatchdog(rootPath string, retention time.Duration) {
	go func() {
		exportDir := filepath.Join(rootPath, "export")
		// Check every 1 hour
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			entries, err := os.ReadDir(exportDir)
			if err != nil {
				// log.Printf("Export Watchdog Error: %v\n", err)
				LOG.Error("Export Watchdog Error", "error", err)
				continue
			}

			cutoffTime := time.Now().Add(-retention)

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				info, err := entry.Info()
				if err != nil {
					continue
				}

				// If the file is older than our retention policy, delete it
				if info.ModTime().Before(cutoffTime) {
					filePath := filepath.Join(exportDir, entry.Name())
					os.Remove(filePath)
				}
			}
		}
	}()
}