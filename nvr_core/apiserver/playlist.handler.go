package apiserver

import (
	"errors"
	"fmt"
	"net/http"

	"nvr_core/service"
	"nvr_core/utils"
)

// HandleGetPlaylist expects: GET /api/cameras/{cam_id}/playlist.m3u8?start=1711000000&end=1711003600
func (api *APIServer) HandleGetPlaylist(w http.ResponseWriter, r *http.Request) {
	camID, idErr := utils.Str2CamID(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	start, end, err := GetMSTimeRange(r)
	if err != nil {
		http.Error(w, "Invalid start or end timestamps", http.StatusBadRequest)
		return
	}

	// Determine the base URL dynamically so it works on localhost, LAN IPs, or reverse proxies
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	profile := GetQueryProfile(r)

	// Call the Service
	playlist, err := api.Services.Playlist.GeneratePlaylist(r.Context(), camID, profile, start, end, baseURL)

	if err != nil {
		if errors.Is(err, service.ErrVideoNotFound) {
			http.Error(w, "No video found for this time range", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Crucial: Set the Apple HTTP Live Streaming MIME type
	// w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Content-Type", "audio/x-mpegurl")
	
	// Prevent caching so the browser/VLC always asks for fresh playlists
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Write the M3U8 string to the client
	w.Write([]byte(playlist))
}