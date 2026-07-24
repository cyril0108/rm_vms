package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/utils"
)

type ExportParams struct {
	TM       *ExportTaskManager `json:"-"`
	TaskID  string    `json:"task_id"`
	CamID    int64    `json:"cam_id"`
	Profile  string   `json:"profile"`
	Format   string   `json:"format"`
	Start    int64    `json:"start"`
	End      int64    `json:"end"`
}


type ExportService interface {
	ExportTimeRange(ctx context.Context, rootPath string, params ExportParams) (string, error)
}

func NewExportService(repo repository.SegmentRepository) ExportService {
	return &segmentServiceBase{repo: repo}
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

func (s *segmentServiceBase) ExportTimeRange(ctx context.Context, rootPath string, params ExportParams) (string, error) {

	format := params.Format

	// Validate Format
	if err := utils.ValidateExportFormat(format); err != nil {
		return "", err
	}

	taskID   := params.TaskID
	camID    := params.CamID
	profile  := params.Profile
	reqStart := params.Start
	reqEnd   := params.End

	params.TM.UpdateTaskStatus(taskID, utils.TaskStatusRunning)

	// Fetch and Sort Segments (Database Responsibility)
	segments, err := s.repo.GetProfileSegmentsByRange(ctx, camID, profile, reqStart, reqEnd)
	if err != nil {
		return "", err
	}
	if len(segments) == 0 {
		return "", fmt.Errorf("no recordings found in the specified time range")
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].StartTime < segments[j].StartTime
	})

	var audioCodec string
	if len(segments) > 0 {
		// Use a short 2-second timeout context just in case ffprobe hangs
		probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		audioCodec = utils.ProbeAudioCodec(probeCtx, segments[0].FilePath)
		cancel()
	}

	// Build the Concat File (Math & File I/O Responsibility)
	concatFilePath, err := s.buildConcatFile(segments, reqStart, reqEnd)
	if err != nil {
		return "", err
	}
	defer os.Remove(concatFilePath) // Ensure cleanup regardless of success/failure

	// Ensure Export Directory Exists
	exportDir := filepath.Join(rootPath, "export")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %v", err)
	}

	// Execute FFmpeg (Subprocess & Codec Responsibility)
	outputPath := filepath.Join(exportDir, fmt.Sprintf("export_cam%d_%d.%s", camID, reqStart, format))

	// Calculate the total duration in seconds for the progress math
	totalDurationSec := float64(params.End - params.Start) / 1000.0

	// Define the Callback: This is the ONLY place that talks to the Task Manager
	progressCallback := func(progress float64) {
		if params.TM != nil {
			params.TM.UpdateTaskProgress(taskID, progress)
		}
	}

	if err := s.executeFFmpeg(ctx, concatFilePath, outputPath, format, audioCodec, totalDurationSec, progressCallback); err != nil {
		return "", err
	}

	return outputPath, nil
}


func (s *segmentServiceBase) buildConcatFile(segments []*models.Segment, reqStart, reqEnd int64) (string, error) {
	concatFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("concat_tasks_%d.txt", reqStart))
	file, err := os.Create(concatFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create concat file: %v", err)
	}
	defer file.Close()

	for i, seg := range segments {
		fmt.Fprintf(file, "file '%s'\n", seg.FilePath)

		// First segment: Trim the beginning
		if i == 0 && reqStart > seg.StartTime {
			offsetSeconds := float64(reqStart-seg.StartTime) / 1000.0
			fmt.Fprintf(file, "inpoint %.3f\n", offsetSeconds)
		}

		// Last segment: Trim the end
		if i == len(segments)-1 && reqEnd < seg.EndTime {
			durationSeconds := float64(reqEnd-seg.StartTime) / 1000.0
			fmt.Fprintf(file, "outpoint %.3f\n", durationSeconds)
		}
	}

	file.Sync()
	return concatFilePath, nil
}

// Pre-compile the regex to find "time=HH:MM:SS.ms" in FFmpeg output
var timeRegex = regexp.MustCompile(`time=([0-9]{2}):([0-9]{2}):([0-9]{2}\.[0-9]+)`)

func (s *segmentServiceBase) executeFFmpeg(ctx context.Context, concatFilePath, outputPath, format, audioCodec string, totalDurationSec float64, onProgress func(float64)) error {

	// Base arguments required for all exports
	ffmpegArgs := []string{
		"-y", 
		"-nostdin", // Prevents background TTY/Stdin from pausing the program
		"-f", "concat",
		"-safe", "0",
		"-i", concatFilePath,
	}

	// Inject dynamic codec strategies (Copy vs Transcode)
	ffmpegArgs = append(ffmpegArgs, s.getCodecStrategy(format, audioCodec)...)

	// Finalize with the output path
	ffmpegArgs = append(ffmpegArgs, outputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)

	/// =====================================
	/// Monitor ffmpeg progress
	/// =====================================
	// Create a pipe to read FFmpeg's stderr (where it prints progress)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not create stderr pipe: %v", err)
	}

	// Start the command asynchronously
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg failed to start: %v", err)
	}

	// Parse the output in real-time
	scanner := bufio.NewScanner(stderr)
	// FFmpeg uses carriage returns (\r) instead of newlines (\n) for progress updates
	scanner.Split(bufio.ScanWords) 

	var lastProgress int = -1

	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, "time=") {
			matches := timeRegex.FindStringSubmatch(text)
			if len(matches) == 4 {
				hours, _ := strconv.ParseFloat(matches[1], 64)
				mins, _ := strconv.ParseFloat(matches[2], 64)
				secs, _ := strconv.ParseFloat(matches[3], 64)

				currentSec := (hours * 3600) + (mins * 60) + secs

				// Calculate percentage
				progress := (currentSec / totalDurationSec) * 100.0
				if progress > 100 {
					progress = 100
				}

				// Throttle updates: Only trigger the callback if the whole integer % changes
				// This protects your TaskManager Mutex from being hammered 100x a second
				currentInt := int(progress)
				if currentInt > lastProgress {
					lastProgress = currentInt
					if onProgress != nil {
						onProgress(progress) // Fire the callback
					}
				}
			}
		// } else {

		// 	LOG.Info("[executeFFmpeg]", "msg", text)

		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg execution failed: %v", err)
	}

	return nil
}

// getCodecStrategy determines if the stream requires re-encoding
func (s *segmentServiceBase) getCodecStrategy(format string, audioCodec string) []string {

	hasAudio := audioCodec != ""
	isG711 := audioCodec == "pcm_alaw" || audioCodec == "pcm_mulaw" || audioCodec == "alaw" || audioCodec == "mulaw"

	// AVI is a legacy container. We force a transcode to x264 to ensure 
	// compatibility, utilizing the "superfast" preset so we don't melt the NVR's CPU.
	if format == "avi" {
		args := []string{
			"-c:v", "libx264",
			"-preset", "superfast",
			"-crf", "23",
		}
		if hasAudio {
			args = append(args, "-c:a", "aac")
		}
		return args
	}

	// Modern formats (MP4, MKV, MOV) get the blazing fast Stream Copy
	args := []string{"-c", "copy"}

	if (format == "mp4" || format == "mov") && isG711 {
		args = []string{
			"-c:v", "copy",   // STILL COPY VIDEO (Critical for NVR performance)
			"-c:a", "aac",    // Transcode audio to AAC
			"-b:a", "128k",   // Standard audio bitrate
		}
	}

	// Apply Faststart only to ISO base media formats
	if format == "mp4" || format == "mov" {
		args = append(args, "-movflags", "+faststart")
	}

	return args
}