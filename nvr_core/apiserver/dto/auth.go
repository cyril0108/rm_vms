package dto

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	Permissions  []string `json:"permissions"`
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