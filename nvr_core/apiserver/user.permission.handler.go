package apiserver

import (
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/apiserver/middleware"
)

func (api *APIServer) HandleGetAllRoles(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	if session, ok := middleware.GetSession(ctx);
		!ok || !session.HasPermissionUserManage() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	roles, err := api.Services.Perms.GetAllRoles(ctx)
	if err != nil {
		http.Error(w, "Failed to get all roles", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, roles)


}

func (api *APIServer) HandleGetAllPermissions(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	if session, ok := middleware.GetSession(ctx);
		!ok || !session.HasPermissionUserManage() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	perms, err := api.Services.Perms.GetAllPermissions(ctx)
	if err != nil {
		http.Error(w, "Failed to get all permissions", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, perms)

}

// HandleUpdateUserPermissions expects: PUT /api/users/{id}/permissions
func (api *APIServer) HandleUpdateUserPermissions(w http.ResponseWriter, r *http.Request) {

	session, ok := middleware.GetSession(r.Context())
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

	var req *dto.UpdatePermissionsRequest
	if err := decodeRequest(r, req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Call the service layer to perform the transactional swap
	err = api.Services.User.UpdateUserPermissions(r.Context(), session.UserID, targetUserID, req.PermissionIDs)
	if err != nil {
		// Log error internally, return generic 500 or 404 to client
		http.Error(w, "Failed to update permissions", http.StatusInternalServerError)
		return
	}

	api.Services.Auth.UpdateUserStatusForPermissionChange(targetUserID)

	w.WriteHeader(http.StatusOK)
	RespondJSON(w, true)
}