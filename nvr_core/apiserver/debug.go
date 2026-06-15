package apiserver

import (
	"net/http"
	"nvr_core/utils"
)

func (s *APIServer) GetDebugInfo(w http.ResponseWriter, r *http.Request) {

	data, error := s.Services.System.GetDebugData(s.Context)

	if(error != nil) {
		utils.RespondJSONHTTPStatus(w, "failed to get debug info.", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, data, "")
}