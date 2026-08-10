package apiserver

import (
	"net/http"

	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
	"nvr_core/db/models"
	"nvr_core/utils"
)

// HandleListEmailGroups returns all recipient groups with their subscribed event types.
func (s *APIServer) HandleListEmailGroups(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	groups, err := s.Services.Email.ListGroupsWithEvents(s.Context)
	if err != nil {
		LOG.Error("Error listing email groups", "error", err)
		utils.RespondJSONHTTPStatus(w, "Error listing email groups", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.EmailGroupResponse, 0, len(groups))
	for _, g := range groups {
		eventTypes := g.EventTypes
		if eventTypes == nil {
			eventTypes = []string{}
		}
		recipients := []string(g.Group.Recipients)
		if recipients == nil {
			recipients = []string{}
		}
		resp = append(resp, dto.EmailGroupResponse{
			ID:         g.Group.ID,
			Name:       g.Group.Name,
			Recipients: recipients,
			EventTypes: eventTypes,
		})
	}

	utils.RespondJSON(w, resp, "success")
}

// HandleCreateEmailGroup creates a new recipient group with optional event type bindings.
func (s *APIServer) HandleCreateEmailGroup(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req dto.EmailGroupRequest
	if err := decodeRequest(r, &req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	if req.Name == "" {
		utils.RespondJSONHTTPStatus(w, "Group name is required", http.StatusBadRequest)
		return
	}

	if len(req.Recipients) == 0 {
		utils.RespondJSONHTTPStatus(w, "At least one recipient is required", http.StatusBadRequest)
		return
	}

	group := &models.EmailGroup{
		Name:       req.Name,
		Recipients: models.EmailRecipients(req.Recipients),
	}

	id, err := s.Services.Email.CreateGroup(s.Context, group, req.EventTypes)
	if err != nil {
		LOG.Error("Error creating email group", "error", err)
		utils.RespondJSONHTTPStatus(w, "Error creating email group", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]int64{"id": id}, "success")
}

// HandleUpdateEmailGroup updates a group's name, recipients, and event type bindings.
func (s *APIServer) HandleUpdateEmailGroup(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	id, err := getPathID(r, "id")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req dto.EmailGroupRequest
	if err := decodeRequest(r, &req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	if req.Name == "" {
		utils.RespondJSONHTTPStatus(w, "Group name is required", http.StatusBadRequest)
		return
	}

	group := &models.EmailGroup{
		ID:         id,
		Name:       req.Name,
		Recipients: models.EmailRecipients(req.Recipients),
	}

	if err := s.Services.Email.UpdateGroup(s.Context, group, req.EventTypes); err != nil {
		LOG.Error("Error updating email group", "error", err)
		utils.RespondJSONHTTPStatus(w, "Error updating email group", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, "", "success")
}

// HandleDeleteEmailGroup removes a group and its event type bindings (cascade).
func (s *APIServer) HandleDeleteEmailGroup(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermission(middleware.PERMSystem) {
		utils.RespondJSONHTTPStatus(w, "Forbidden", http.StatusForbidden)
		return
	}

	id, err := getPathID(r, "id")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	if err := s.Services.Email.DeleteGroup(s.Context, id); err != nil {
		LOG.Error("Error deleting email group", "error", err)
		utils.RespondJSONHTTPStatus(w, "Error deleting email group", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, "", "success")
}
