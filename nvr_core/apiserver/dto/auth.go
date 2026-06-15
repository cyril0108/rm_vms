package dto

import "nvr_core/db/models"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginUser struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	RoleID   int64  `json:"role_id"`
	Email    string `json:"email"`
	// IsActive bool   `json:"is_active"`
}

type LoginResponse struct {
	User         LoginUser `json:"user"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Permissions  []string  `json:"permissions"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshLogoutRequest struct {
	Logout bool `json:"logout"`
}

type RefreshResponse struct {
	Token       string   `json:"token"`
	Permissions []string `json:"permissions"`
}

func NewLoginUser(user *models.User) *LoginUser {
	return &LoginUser{
		UserID: user.ID,
		Username: user.Username,
		Name: user.Name,
		RoleID: user.RoleID,
		Email: user.Email,
	}
}