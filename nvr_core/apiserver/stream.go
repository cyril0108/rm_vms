package apiserver

import (
	// "context"
	// "encoding/json"
	"log"
	"net/http"

	// "sync"
	// "time"
	"nvr_core/stream"
	"nvr_core/utils"
)

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetStream(w http.ResponseWriter, r *http.Request) {

	id, idErr := utils.Str2Int(r.PathValue("cam_id"))
	if(idErr != nil) {
		http.Error(w, "Invalid cam id", http.StatusBadRequest)
		return
	}

	profile := utils.SanitizeProfile(r.PathValue("profile"))

	worker, err := s.PM.CameraWorker(id, profile)
	if(err != nil) {
		log.Println("[GetStream] failed to get worker")
		return
	}

	hub := worker.StreamHubForCam(id, profile)
	if(hub == nil) {
		log.Println("[GetStream] failed to get stream hub")
		return
	}

	stream.ServeWs(hub, w, r)

}
