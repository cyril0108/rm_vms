package apiserver

import (
	"log"
	"net/http"
)

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {

	data, err := s.Services.System.GetDebugData(s.Context)
	if err != nil {
		http.Error(w, "failed to get health info.", http.StatusInternalServerError)
		return
	}

	if err := RespondJSON(w, data); err != nil {
		log.Printf("Error checking health: %v", err)
	}

}
