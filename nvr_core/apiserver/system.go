package apiserver

import (
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
	"nvr_core/buildinfo"
	"nvr_core/hardware"
	"nvr_core/utils"
)

// HandleAdminInitConfigure
//
func (s *APIServer) HandleGetMachineInfo(w http.ResponseWriter, r *http.Request) {

	info := dto.SystemMachineInfo{
		MachineID: hardware.GetPersistentMachineID(),
	}

	name, err := s.Services.SysSetting.GetServerName(s.Context)
	if err != nil {
		LOG.Error("Error fetching server name", "error", err)
	}

	info.ServerName = name

	info.Version = buildinfo.Version

	utils.RespondJSON(w, info, "success")
	return
}


func (s *APIServer) HandleSetServerName(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req dto.SystemSettingRequest
	if err := decodeRequest(r, &req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	if req.Value == "" {
		utils.RespondJSONHTTPStatus(w, "Server name cannot be empty", http.StatusBadRequest)
		return
	}

	err := s.Services.SysSetting.SetServerName(s.Context, req.Value)
	if err != nil {
		LOG.Error("Error setting server name", "error", err)
		utils.RespondJSONHTTPStatus(w, "Error setting server name", http.StatusInternalServerError)
	}

	utils.RespondJSON(w, "", "success")
	return
}


