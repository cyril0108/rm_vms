package service

import (
	"bufio"
	"context"
	"fmt"
	"image"
	_ "image/png"
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
	TM        *ExportTaskManager `json:"-"`
	TaskID    string             `json:"task_id"`
	CamID     int64              `json:"cam_id"`
	Profile   string             `json:"profile"`
	Format    string             `json:"format"`
	Start     int64              `json:"start"`
	End       int64              `json:"end"`
	Watermark *WatermarkParams   `json:"-"`
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

	sampleSegPath := ""
	if len(segments) > 0 {
		sampleSegPath = segments[0].FilePath
	}

	if err := s.executeFFmpeg(ctx, concatFilePath, outputPath, format, audioCodec, totalDurationSec, params.Watermark, sampleSegPath, progressCallback); err != nil {
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

func (s *segmentServiceBase) executeFFmpeg(ctx context.Context, concatFilePath, outputPath, format, audioCodec string, totalDurationSec float64, wm *WatermarkParams, sampleSegPath string, onProgress func(float64)) error {

	// Base arguments required for all exports
	ffmpegArgs := []string{
		"-y", 
		"-nostdin", // Prevents background TTY/Stdin from pausing the program
		"-f", "concat",
		"-safe", "0",
		"-i", concatFilePath,
	}

	if wm != nil && (wm.Text != "" || wm.ImagePath != "") {
		hasImage := wm.ImagePath != ""
		hasText := wm.Text != ""

		var videoW, videoH, pngW, pngH int
		if hasImage {
			ffmpegArgs = append(ffmpegArgs, "-i", wm.ImagePath)
			probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			var err error
			videoW, videoH, err = probeVideoDimensions(probeCtx, sampleSegPath)
			cancel()
			if err != nil {
				return fmt.Errorf("probe video failed: %v", err)
			}
			imgW, imgH, err := getImageDimensions(wm.ImagePath)
			if err != nil {
				return fmt.Errorf("read image failed: %v", err)
			}
			if imgW > 0 {
				pngW = videoW * wm.Scale / 100
				pngH = imgH * pngW / imgW
				if pngW < 1 {
					pngW = 1
				}
				if pngH < 1 {
					pngH = 1
				}
			}
		}

		var textFilePath string
		if hasText {
			tf, err := os.CreateTemp("", "wm_text_*.txt")
			if err != nil {
				return err
			}
			defer os.Remove(tf.Name())
			tf.WriteString(wm.Text)
			tf.Close()
			textFilePath = tf.Name()
		}

		filter := buildWatermarkFilter(*wm, hasImage, hasText, videoW, videoH, pngW, pngH, textFilePath)

		if hasImage {
			ffmpegArgs = append(ffmpegArgs, "-filter_complex", filter, "-map", "[out]", "-map", "0:a?")
		} else {
			ffmpegArgs = append(ffmpegArgs, "-vf", filter)
		}

		ffmpegArgs = append(ffmpegArgs, "-c:v", "libx264", "-preset", "superfast", "-crf", "23")

		isG711 := audioCodec == "pcm_alaw" || audioCodec == "pcm_mulaw" || audioCodec == "alaw" || audioCodec == "mulaw"
		if audioCodec != "" {
			if (format == "mp4" || format == "mov") && isG711 {
				ffmpegArgs = append(ffmpegArgs, "-c:a", "aac", "-b:a", "128k")
			} else {
				ffmpegArgs = append(ffmpegArgs, "-c:a", "copy")
			}
		}

		if format == "mp4" || format == "mov" {
			ffmpegArgs = append(ffmpegArgs, "-movflags", "+faststart")
		}
	} else {
		// Inject dynamic codec strategies (Copy vs Transcode)
		ffmpegArgs = append(ffmpegArgs, s.getCodecStrategy(format, audioCodec)...)
	}

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

// --- Watermark ---

type WatermarkParams struct {
	InputPath string
	Text      string
	ImagePath string
	Position  string
	Scale     int
	Opacity   int
	Color     string
}

const (
	wmPadding  = 10
	wmGap      = 5
	wmFontSize = 24
)



func buildWatermarkFilter(p WatermarkParams, hasImage, hasText bool, videoW, videoH, pngW, pngH int, textFilePath string) string {
	opacity := float64(p.Opacity) / 100.0
	rgb, textAlpha := parseRGBAColor(p.Color)
	fontOpt := findFontOpt()

	if hasImage && hasText {
		ox, oy := overlayPosition(p.Position, videoW, videoH, pngW, pngH, true)
		tx, ty := textPositionBelow(p.Position, oy, pngH)
		return fmt.Sprintf(
			"[1:v]scale=%d:%d,format=rgba,colorchannelmixer=aa=%.2f[wm];"+
				"[0:v][wm]overlay=%d:%d,"+
				"drawtext=%stextfile='%s':fontsize=%d:fontcolor=0x%s@%.2f:x=%s:y=%s[out]",
			pngW, pngH, opacity,
			ox, oy,
			fontOpt, textFilePath, wmFontSize, rgb, textAlpha, tx, ty,
		)
	}

	if hasImage {
		ox, oy := overlayPosition(p.Position, videoW, videoH, pngW, pngH, false)
		return fmt.Sprintf(
			"[1:v]scale=%d:%d,format=rgba,colorchannelmixer=aa=%.2f[wm];"+
				"[0:v][wm]overlay=%d:%d[out]",
			pngW, pngH, opacity,
			ox, oy,
		)
	}

	tx, ty := textPosition(p.Position)
	return fmt.Sprintf(
		"drawtext=%stextfile='%s':fontsize=%d:fontcolor=0x%s@%.2f:x=%s:y=%s",
		fontOpt, textFilePath, wmFontSize, rgb, textAlpha, tx, ty,
	)
}

func overlayPosition(position string, videoW, videoH, pngW, pngH int, withText bool) (int, int) {
	textSpace := 0
	if withText {
		textSpace = wmFontSize + wmGap
	}
	totalH := pngH + textSpace

	switch position {
	case "top-right":
		return videoW - pngW - wmPadding, wmPadding
	case "center":
		return (videoW - pngW) / 2, (videoH - totalH) / 2
	case "bottom-left":
		return wmPadding, videoH - totalH - wmPadding
	case "bottom-right":
		return videoW - pngW - wmPadding, videoH - totalH - wmPadding
	default:
		return wmPadding, wmPadding
	}
}

func textPositionBelow(position string, overlayY, pngH int) (string, string) {
	ty := fmt.Sprintf("%d", overlayY+pngH+wmGap)
	switch position {
	case "top-right", "bottom-right":
		return fmt.Sprintf("w-tw-%d", wmPadding), ty
	case "center":
		return "(w-tw)/2", ty
	default:
		return fmt.Sprintf("%d", wmPadding), ty
	}
}

func textPosition(position string) (string, string) {
	pad := fmt.Sprintf("%d", wmPadding)
	switch position {
	case "top-right":
		return "w-tw-" + pad, pad
	case "center":
		return "(w-tw)/2", "(h-th)/2"
	case "bottom-left":
		return pad, "h-th-" + pad
	case "bottom-right":
		return "w-tw-" + pad, "h-th-" + pad
	default:
		return pad, pad
	}
}

func probeVideoDimensions(ctx context.Context, path string) (int, int, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected ffprobe output: %s", out)
	}
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])
	return w, h, nil
}

func getImageDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func parseRGBAColor(color string) (string, float64) {
	if len(color) != 8 {
		return "FFFFFF", 0.25
	}
	rgb := color[:6]
	aa, err := strconv.ParseUint(color[6:8], 16, 8)
	if err != nil {
		return rgb, 0.25
	}
	return rgb, float64(aa) / 255.0
}

func findFontOpt() string {
	paths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return "fontfile=" + p + ":"
		}
	}
	return ""
}