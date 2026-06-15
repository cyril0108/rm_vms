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

	duration, _ := GetDurationTime(r)
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