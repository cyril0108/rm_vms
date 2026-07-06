package apiserver

import (
	"encoding/json"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
	"nvr_core/utils"
)

// HandleAdminInitConfigure
func (s *APIServer) HandleAdminInitConfigure(w http.ResponseWriter, r *http.Request) {

	ctx := s.Context

	if health, err := s.Services.System.GetHealthData(ctx); err != nil || health.Configured {
		utils.RespondJSONHTTPStatus(w, "System already configured.", http.StatusConflict)
		return;
	}

	var req dto.SystemConfigureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	if req.Username == "" || req.Password == "" {
		utils.RespondJSONHTTPStatus(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	u := &models.User {
		Username: req.Username,
		Password: req.Password,
		RoleID: 1,
		IsActive: true,
	}

	if err := s.Services.User.CreateUser(ctx, 1, u); err != nil {
		LOG.Error("Error when creating admin", "err", err)
		utils.RespondJSONHTTPStatus(w, "Failed to configure admin", http.StatusInternalServerError)
		return
	}

	if req.ServerName != "" {

		if err := s.Services.SysSetting.SetServerName(ctx, req.ServerName); err != nil {
			LOG.Error("Error when setting server name", "err", err)
			utils.RespondJSONHTTPStatus(w, "Error when setting server name", http.StatusInternalServerError)
			return
		}

	}

	utils.RespondJSON(w, "", "success")
	return
}
