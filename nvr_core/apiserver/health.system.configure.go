package apiserver

import (
	"encoding/json"
	"net/http"

	"nvr_core/apiserver/dto"
)

// HandleAdminInitConfigure
// 
func (s *APIServer) HandleAdminInitConfigure(w http.ResponseWriter, r *http.Request) {

	ctx := s.Context

	if health, err := s.Services.System.GetHealthData(ctx); err != nil || health.Configured {
		http.Error(w, "System already configured.", http.StatusConflict)
		return;
	}

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := s.Services.User.UpdateUserPassword(ctx, 1, 1, req.Password); err != nil {
		LOG.Error("Error when updating admin password %v", err)
		http.Error(w, "Failed to configure admin password", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, "success")
	return
}
