package apiserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	// "nvr_core/apiserver/dto"
	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
)




// HandleListUsers expects: GET /api/admin/users?page=1&limit=20
func (api *APIServer) HandleListUsers(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
	if !ok || !session.HasPermissionUserManage() { // Strictly enforce permissions
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse pagination query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	users, total, err := api.Services.User.GetAllUsers(r.Context(), page, limit)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]interface{}{
		"users": users,
		"total": total,
		"page":  page,
	})
}

// HandleDeactivateUser expects: POST /api/admin/users/create
func (api *APIServer) HandleCreateUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok || !session.HasPermissionUserManage() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req dto.CreateUserRequest
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	newUser := req.MapToDBUser()
	if err := api.Services.User.CreateUser(ctx, session.UserID, newUser); err != nil {
		LOG.Info("Failed to create user", "err", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]int64 {
		"id": newUser.ID,
	})

}

// HandleDeactivateUser expects: DELETE /api/admin/users/{id}
func (api *APIServer) HandleDeactivateUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = api.Services.User.DeactivateUser(ctx, session.UserID, targetUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Important: Remember to instantly revoke all active refresh tokens for this user!
	_ = api.Services.Auth.LogoutDeactivatedUser(ctx, targetUserID)

	RespondJSON(w, map[string]string{"message": "User deactivated successfully"})
}

// HandleUpdateUserRole expects: PATCH /api/admin/users/{id}/role
func (api *APIServer) HandleUpdateUserRole(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	session, ok := middleware.GetSession(ctx)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	targetUserID, err := getPathID(r, "id")
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if !session.HasPermissionUserNoSelfManage(targetUserID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var payload struct {
		RoleID int64 `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	err = api.Services.User.UpdateUserRole(ctx, session.UserID, targetUserID, payload.RoleID)
	if err != nil {
		http.Error(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	// Instantly expire their current session so they are forced to refresh and get their new permissions
	api.Services.Auth.UpdateUserStatusForPermissionChange(targetUserID)

	RespondJSON(w, map[string]string{"message": "Role updated successfully"})
}