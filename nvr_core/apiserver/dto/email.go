package dto

// ──────────────────────────────────────────────
// SMTP Settings
// ──────────────────────────────────────────────

type EmailSMTPSettingsRequest struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`     // empty string = keep existing password
	SenderEmail string `json:"sender_email"`
	SenderName  string `json:"sender_name"`
	UseTLS      bool   `json:"use_tls"`
	Enabled     bool   `json:"enabled"`
}

type EmailSMTPSettingsResponse struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`     // always masked as "********"
	SenderEmail string `json:"sender_email"`
	SenderName  string `json:"sender_name"`
	UseTLS      bool   `json:"use_tls"`
	Enabled     bool   `json:"enabled"`
}

// ──────────────────────────────────────────────
// Email Groups
// ──────────────────────────────────────────────

type EmailGroupRequest struct {
	Name       string   `json:"name"`
	Recipients []string `json:"recipients"`
	EventTypes []string `json:"event_types"`
}

type EmailGroupResponse struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Recipients []string `json:"recipients"`
	EventTypes []string `json:"event_types"`
}

// ──────────────────────────────────────────────
// Test Email
// ──────────────────────────────────────────────

type EmailTestRequest struct {
	To string `json:"to"`
}
