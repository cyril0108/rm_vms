package apiserver

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"

	"nvr_core/service"
	"nvr_core/utils"
)

// HandlePlayVideo expects: GET /api/cameras/{id}/play/ts?profile=sub&time=1711000050
func (api *APIServer) HandleTransmuxTS(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	timestamp, err := GetSearchAtTime(r)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	duration, err := GetDurationTime(r)
	if duration == 0 && err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	profile := GetQueryProfile(r)

	// Get the validated physical path from the Service
	seg, err := api.Services.Playback.GetVideoSegment(r.Context(), camID, profile, timestamp)
	if err != nil {
		if errors.Is(err, service.ErrVideoSegmentNotFound) || errors.Is(err, service.ErrFileMissing) {
			utils.RespondJSONHTTPStatus(w, "Video not found", http.StatusNotFound)
			return
		}
		utils.RespondJSONHTTPStatus(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/MP2T")

	// Add headers to prevent caching of video streams (crucial for NVRs)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")


	// Create ffmpeg arguments
	skip := utils.NewFFTime((timestamp - seg.StartTime))

	ffmpegArgs := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", seg.FilePath,
	}

	// If skip > 0, inject -ss BEFORE the input file for fast seeking
	if skip.MS > 0 {
	    ffmpegArgs = append([]string{"-ss", skip.TimeString()}, ffmpegArgs...)
	}

	// If duration exists, inject -t to stop transcoding at the exact end time
	if duration > 0 {
		d := utils.NewFFTime(duration)
	    ffmpegArgs = append(ffmpegArgs, "-t", d.TimeString())
	}

	// Standard TS output arguments
	ffmpegArgs = append(ffmpegArgs,
		"-c:v", "copy",                       
		"-bsf:v", "h264_mp4toannexb",         
		"-c:a", "aac", // Replace -an with this if Safari stays black
		"-b:a", "64k",
		"-f", "mpegts",                       
		"-muxdelay", "0",                     // Removes pipe buffering latency
		"pipe:1",                             
	)

	// Using the r.Context() ensures FFmpeg dies if the client disconnects
	// cmd := exec.CommandContext(r.Context(), "ffmpeg",
	// 	"-hide_banner", "-loglevel", "error", // Suppress noisy logs
	// 	"-i", filePath,                       // Input the MKV
	// 	"-c:v", "copy",                       // Zero-CPU Video Copy
	// 	"-an",                                // Drop incompatible audio (Change to "-c:a aac" if you want audio)
	// 	"-f", "mpegts",                       // Force MPEG-TS format
	// 	"pipe:1",                             // Output to stdout instead of a file
	// )

	cmd := exec.CommandContext(r.Context(), "ffmpeg", ffmpegArgs...)
	// cmd := exec.CommandContext(r.Context(), "ffmpeg",
	// 		"-hide_banner", "-loglevel", "error", 
	// 		"-i", seg.FilePath,                       
	// 		"-c:v", "copy",                       
	// 		"-bsf:v", "h264_mp4toannexb",         
	// 		"-c:a", "aac", // Replace -an with this if Safari stays black
	// 		"-b:a", "64k",
	// 		"-f", "mpegts",                       
	// 		"-muxdelay", "0",                     // Removes pipe buffering latency
	// 		"pipe:1",                             
	// 	)

	// Connect FFmpeg's stdout directly to the HTTP Response Writer
	cmd.Stdout = w

	// Optional: Pipe stderr to your Go logger to catch FFmpeg issues
	cmd.Stderr = os.Stderr 

	// Execute and stream
	if err := cmd.Run(); err != nil {
		// If the context was canceled (client disconnected), this is expected.
		if r.Context().Err() != nil {
			return
		}
		log.Printf("[playback.ts.handler] FFmpeg error for %s: %v", seg.FilePath, err)
	}

}


// HandleGapFillerTS expects: GET /api/cameras/{cam_id}/play/gap?duration=5000 (or however you define it)
func (api *APIServer) HandleGapFillerTS(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		utils.RespondJSONHTTPStatus(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	cam := api.PM.GetCamera(int(camID))
	if cam == nil {
		utils.RespondJSONHTTPStatus(w, "No camera data with given ID", http.StatusNotFound)
		return
	}

	hasAudio := cam.SubStream.ACodec != 0

	// Get duration for the gap (you can use your existing helper)
	duration, err := GetDurationTime(r)
	if duration == 0 || err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid or missing duration", http.StatusBadRequest)
		return
	}

	// Set the exact same headers so the browser treats it as a normal video segment
	w.Header().Set("Content-Type", "video/MP2T")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Convert duration to FFmpeg time format using your existing util
	d := utils.NewFFTime(duration)

	// Build FFmpeg arguments for a synthetic black screen + silent audio
	ffmpegArgs := []string{
		"-hide_banner", "-loglevel", "error",

		// --- Inputs ---
		// Generate pure black video (adjust resolution 's' to match your sub/main profiles if needed)
		"-f", "lavfi", "-i", "color=c=black:s=1280x720:r=15",
	}

	if hasAudio {
		ffmpegArgs = append(ffmpegArgs, "-f", "lavfi", "-i", "anullsrc=channel_layout=stereo:sample_rate=44100") // Input 1: Audio
	}

	// Append duration and video encoding settings
	ffmpegArgs = append(ffmpegArgs,
		"-t", d.TimeString(),
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "stillimage",
		"-pix_fmt", "yuv420p",
	)

	// Conditionally add audio encoding settings
	if hasAudio {
		ffmpegArgs = append(ffmpegArgs,
			"-c:a", "aac",
			"-b:a", "64k",
		)
	}

	// Append output format and piping
	ffmpegArgs = append(ffmpegArgs,
		"-f", "mpegts",
		"-muxdelay", "0",
		"pipe:1",
	)

	// Execute and pipe directly to HTTP response
	cmd := exec.CommandContext(r.Context(), "ffmpeg", ffmpegArgs...)
	
	cmd.Stdout = w
	cmd.Stderr = os.Stderr // Pipe stderr to catch encoder failures

	if err := cmd.Run(); err != nil {
		// Context canceled means the client scrubbed away or disconnected
		if r.Context().Err() != nil {
			return
		}
		log.Printf("[playback.ts.handler] Gap filler FFmpeg error: %v", err)
	}
}