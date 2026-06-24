package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
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