package apiserver

import (
	"fmt"
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
	"nvr_core/db/models"
	"nvr_core/utils"
)

// HandleGetSMTPSettings returns the current SMTP configuration with the password masked.
func (s *APIServer) HandleGetSMTPSettings(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	settings, err := s.Services.Email.GetSMTPSettings(s.Context)
	if err != nil {
		LOG.Error("Error fetching SMTP settings", "error", err)
		utils.RespondJSONHTTPStatus(w, "Error fetching SMTP settings", http.StatusInternalServerError)
		return
	}

	resp := dto.EmailSMTPSettingsResponse{
		Host:        settings.Host,
		Port:        settings.Port,
		Username:    settings.Username,
		Password:    "********",
		SenderEmail: settings.SenderEmail,
		SenderName:  settings.SenderName,
		UseTLS:      settings.UseTLS,
		Enabled:     settings.Enabled,
	}

	utils.RespondJSON(w, resp, "success")
}

// HandleUpdateSMTPSettings creates or updates the SMTP configuration.
// An empty password field preserves the existing password.
func (s *APIServer) HandleUpdateSMTPSettings(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req dto.EmailSMTPSettingsRequest
	if err := decodeRequest(r, &req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	settings := &models.EmailSMTPSettings{
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		Password:    req.Password,
		SenderEmail: req.SenderEmail,
		SenderName:  req.SenderName,
		UseTLS:      req.UseTLS,
		Enabled:     req.Enabled,
	}

	if err := s.Services.Email.UpdateSMTPSettings(s.Context, settings); err != nil {
		LOG.Error("Error updating SMTP settings", "error", err)
		utils.RespondJSONHTTPStatus(w, "Error updating SMTP settings", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, "", "success")
}

// HandleTestEmail sends a test email using the current SMTP configuration.
func (s *APIServer) HandleTestEmail(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req dto.EmailTestRequest
	if err := decodeRequest(r, &req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	if req.To == "" {
		utils.RespondJSONHTTPStatus(w, "Recipient email is required", http.StatusBadRequest)
		return
	}

	if err := s.Services.Email.SendTestEmail(s.Context, req.To); err != nil {
		LOG.Error("Error sending test email", "error", err)
		utils.RespondJSONHTTPStatus(w, fmt.Sprintf("Failed to send test email: %v", err), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, "", "success")
}
