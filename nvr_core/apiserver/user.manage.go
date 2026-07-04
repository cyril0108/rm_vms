package apiserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	// "nvr_core/apiserver/dto"
	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
	"nvr_core/utils"
)

func (api *APIServer) HandleGetLoginUser(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionUserManage() { // Strictly enforce permissions
		utils.RespondErrForbidden(w)
		return
	}

	user, err := api.Services.User.GetByID(api.Context, session.UserID)

	if err != nil {
		utils.RespondJSONHTTPStatus(w, "failed to get user data", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, user, "success")

}


// HandleListUsers expects: GET /api/admin/users?page=1&limit=20
func (api *APIServer) HandleListUsers(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionUserManage() { // Strictly enforce permissions
		utils.RespondErrForbidden(w)
		return
	}

	// Parse pagination query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	users, total, err := api.Services.User.GetAllUsers(r.Context(), page, limit)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"users": users,
		"total": total,
		"page":  page,
	}, "")
}

// HandleCreateUser expects: POST /api/admin/users/create
func (api *APIServer) HandleCreateUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok || !session.HasPermissionUserManage() {
		utils.RespondErrForbidden(w)
		return
	}

	var req dto.CreateUserRequest
	if err := decodeRequest(r, &req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	newUser := req.MapToDBUser()
	if err := api.Services.User.CreateUser(ctx, session.UserID, newUser); err != nil {
		LOG.Info("Failed to create user", "err", err)
		utils.RespondJSONHTTPStatus(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]int64 {
		"id": newUser.ID,
	}, "")

}

// HandleUpdateUser expects: PUT /api/admin/users/{id}
func (api *APIServer) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) {
		utils.RespondErrForbidden(w)
		return
	}

	var payload dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	maps := payload.MapToPartialInterface()
	err = api.Services.User.UpdatePartial(ctx, session.UserID, targetUserID, maps)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	if _, ok := (maps)["role_id"].(int64); ok {
		// Instantly expire their current session so they are forced to refresh and get their new permissions
		api.Services.Auth.UpdateUserStatusForPermissionChange(targetUserID)
	}

	utils.RespondJSON(w, "", "Role updated successfully")
}

// HandleDeactivateUser expects: DELETE /api/admin/users/{id}
func (api *APIServer) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) {
		utils.RespondErrForbidden(w)
		return
	}

	err = api.Services.User.DeleteUser(ctx, session.UserID, targetUserID)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Important: Remember to instantly revoke all active refresh tokens for this user!
	_ = api.Services.Auth.LogoutDeactivatedUser(ctx, targetUserID)

	utils.RespondJSON(w, "", "User deactivated successfully")
}


// HandleDeactivateUser expects: DELETE /api/admin/users/{id}
func (api *APIServer) HandleDeactivateUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) {
		utils.RespondErrForbidden(w)
		return
	}

	err = api.Services.User.DeactivateUser(ctx, session.UserID, targetUserID)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Important: Remember to instantly revoke all active refresh tokens for this user!
	_ = api.Services.Auth.LogoutDeactivatedUser(ctx, targetUserID)

	utils.RespondJSON(w, "", "User deactivated successfully")
}

func (api *APIServer) HandleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) || targetUserID == session.UserID {
		utils.RespondErrForbidden(w)
		return
	}

	var payload struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = api.Services.User.UpdateUserPassword(ctx, session.UserID, targetUserID, payload.Password)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	// Instantly expire their current session so they are forced to refresh and get their new permissions
	api.Services.Auth.UpdateUserStatusForPermissionChange(targetUserID)

	utils.RespondJSON(w, "", "Role updated successfully")
}

// HandleUpdateUserRole expects: PUT /api/admin/users/{id}/role
func (api *APIServer) HandleUpdateUserRole(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok {
		utils.RespondErrForbidden(w)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) {
		utils.RespondErrForbidden(w)
		return
	}

	var payload struct {
		RoleID int64 `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	err = api.Services.User.UpdateUserRole(ctx, session.UserID, targetUserID, payload.RoleID)
	if err != nil {
		utils.RespondJSONHTTPStatus(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	// Instantly expire their current session so they are forced to refresh and get their new permissions
	api.Services.Auth.UpdateUserStatusForPermissionChange(targetUserID)

	utils.RespondJSON(w, "", "Role updated successfully")
}