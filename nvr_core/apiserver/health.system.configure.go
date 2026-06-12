package apiserver

import (
	"encoding/json"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
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

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	u := &models.User {
		Username: req.Username,
		Password: req.Password,
		RoleID: 1,
	}

	if err := s.Services.User.CreateUser(ctx, 1, u); err != nil {
		LOG.Error("Error when creating admin", "err", err)
		http.Error(w, "Failed to configure admin", http.StatusInternalServerError)
		return
	}


	// if err := s.Services.User.UpdateUserPassword(ctx, 1, 1, req.Password); err != nil {
	// 	LOG.Error("Error when updating admin password %v", err)
	// 	http.Error(w, "Failed to configure admin password", http.StatusInternalServerError)
	// 	return
	// }

	RespondJSON(w, "success")
	return
}
