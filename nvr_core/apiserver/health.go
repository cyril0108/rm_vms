package apiserver

import (
	"log"
	"net/http"
	"nvr_core/utils"
)

// GetCameras safely iterates over the sync.Map
func (s *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {

	data, err := s.Services.System.GetHealthData(s.Context)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get health info.", http.StatusInternalServerError)
		return
	}

	if err := utils.RespondJSON(w, data, ""); err != nil {
		log.Printf("Error checking health: %v", err)
		utils.RespondErrFailedToEncodeResponse(w)
	}

}
