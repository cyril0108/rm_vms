package dto

type SystemConfigureRequest struct {
	ServerName string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

