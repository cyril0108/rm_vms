package apiserver

import (
	"errors"
	"net/http"
	"strconv"

	"nvr_core/service"
	"nvr_core/utils"
)

// HandlePlayVideo expects: GET /api/cameras/{id}/play?time=1711000050
func (api *APIServer) HandlePlayVideo(w http.ResponseWriter, r *http.Request) {

	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	timeStr := r.URL.Query().Get("time")
	timestamp, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid timestamp", http.StatusBadRequest)
		return
	}

	profile := GetQueryProfile(r)

	// Get the validated physical path from the Service
	filePath, err := api.Services.Playback.GetVideoFilePath(r.Context(), camID, profile, timestamp)
	if err != nil {
		if errors.Is(err, service.ErrVideoNotFound) || errors.Is(err, service.ErrFileMissing) {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add headers to prevent caching of video streams (crucial for NVRs)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve the file!
	// http.ServeFile automatically reads the file from disk in chunks, 
	// sets the correct "Content-Type: video/mp4", and natively handles 
	// HTTP 206 Partial Content (Range Requests) so the Vue.js video player can seek.
	http.ServeFile(w, r, filePath)
}