package dto

// ===========================
// User Manage
// ===========================


// ===========================
// Permissions
// ===========================
type UpdatePermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}