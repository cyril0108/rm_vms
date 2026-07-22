package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FFprobeOutput maps the JSON response from ffprobe
type FFprobeOutput struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

// GetVideoResolution inspects a video file and returns its resolution (e.g., "1920x1080").
func GetVideoResolution(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",                          // Suppress logs
		"-select_streams", "v:0",               // Target only the first video stream
		"-show_entries", "stream=width,height", // Extract only width and height
		"-of", "json",                          // Output as JSON
		filePath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffprobe failed on %s: %w", filePath, err)
	}

	var probeData FFprobeOutput
	if err := json.Unmarshal(out.Bytes(), &probeData); err != nil {
		return "", fmt.Errorf("failed to parse ffprobe json: %w", err)
	}

	// Ensure we actually found a video stream with valid dimensions
	if len(probeData.Streams) == 0 || probeData.Streams[0].Width == 0 {
		return "", fmt.Errorf("no valid video stream found in file: %s", filePath)
	}

	res := fmt.Sprintf("%dx%d", probeData.Streams[0].Width, probeData.Streams[0].Height)
	return res, nil
}

// // getVideoDurationMs uses ffprobe to get the actual duration of a video file.
// func GetVideoDurationMs(filePath string) (int64, error) {
// 	// Command: ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 <file>
// 	cmd := exec.Command("ffprobe",
// 		"-v", "error",
// 		"-show_entries", "format=duration",
// 		"-of", "default=noprint_wrappers=1:nokey=1",
// 		filePath,
// 	)

// 	var out bytes.Buffer
// 	cmd.Stdout = &out

// 	if err := cmd.Run(); err != nil {
// 		return 0, fmt.Errorf("ffprobe failed on %s: %w", filePath, err)
// 	}

// 	// Parse the output (e.g., "5.033000")
// 	durationStr := strings.TrimSpace(out.String())
// 	if durationStr == "" {
// 		return 0, fmt.Errorf("empty duration returned by ffprobe for %s", filePath)
// 	}

// 	durationSeconds, err := strconv.ParseFloat(durationStr, 64)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to parse duration float %s: %w", durationStr, err)
// 	}

// 	// Convert to milliseconds
// 	durationMs := int64(durationSeconds * 1000)
// 	return durationMs, nil
// }

// getRealVideoDurationMs calculates duration by counting physical frames and dividing by FPS,
// ignoring corrupted PTS/DTS container metadata.
func GetRealVideoDurationMs(filePath string) (int64, error) {
	// Command: ffprobe -v error -select_streams v:0 -count_packets 
	// -show_entries stream=nb_read_packets,r_frame_rate -of csv=p=0 <file>
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0", // Only look at the video stream
		"-count_packets",         // Force a physical packet count (takes slightly longer, but accurate)
		"-show_entries", "stream=nb_read_packets,r_frame_rate",
		"-of", "csv=p=0",         // Output clean CSV without headers
		filePath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe failed on %s: %w", filePath, err)
	}

	// Output format will be: "r_frame_rate,nb_read_packets" -> e.g., "30/1,1500"
	output := strings.TrimSpace(out.String())
	parts := strings.Split(output, ",")
	if len(parts) != 2 {
		return 0, fmt.Errorf("unexpected ffprobe output format: %s", output)
	}

	frameRateStr := parts[0]
	frameCountStr := parts[1]

	// Parse Frame Rate (usually in format "num/den" like "30/1" or "60000/1001")
	fpsParts := strings.Split(frameRateStr, "/")
	if len(fpsParts) != 2 {
		return 0, fmt.Errorf("unexpected framerate format: %s", frameRateStr)
	}

	fpsNum, errNum := strconv.ParseFloat(fpsParts[0], 64)
	fpsDen, errDen := strconv.ParseFloat(fpsParts[1], 64)
	if errNum != nil || errDen != nil || fpsDen == 0 {
		return 0, fmt.Errorf("invalid framerate values: %s", frameRateStr)
	}
	fps := fpsNum / fpsDen

	// Parse Packet Count
	frameCount, err := strconv.ParseFloat(frameCountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse packet count: %w", err)
	}

	// Calculate Real Duration
	if fps == 0 {
		return 0, fmt.Errorf("calculated fps is 0, cannot divide")
	}

	durationSeconds := frameCount / fps
	durationMs := int64(durationSeconds * 1000)

	return durationMs, nil
}


// ProbeAudioCodec uses ffprobe to quickly extract the audio codec of a media file.
// It returns "pcm_alaw", "pcm_mulaw", "aac", etc.
func ProbeAudioCodec(ctx context.Context, filePath string) string {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	out, err := cmd.Output()
	if err != nil {
		// If probe fails (e.g., video has no audio track), return empty string safely
		return ""
	}

	return strings.TrimSpace(string(out))
}