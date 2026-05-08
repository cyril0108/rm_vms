package models

const CameraTypeONVIF = "onvif"
const CameraTypeRTSP  = "rtsp"

// Camera represents an IP camera (ONVIF or generic RTSP)
type Camera struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`

	Manufacturer          string `json:"manufacturer"`
	Model                 string `json:"model"`
	SerialNumber          string  `json:"serial_number"`

	IPAddress             string  `json:"ip_address"`
	HTTPPort              int     `json:"http_port"`
	Type                  string  `json:"type"`

	Username              string `json:"username"`
	PasswordEnc           string `json:"-"` // CRITICAL: Never serialize to JSON

	StreamURL             string  `json:"stream_url"`
	SubStreamURL          string `json:"sub_stream_url"`

	OnvifProfileToken     string `json:"onvif_profile_token"`
	SubStreamProfileToken string `json:"sub_stream_profile_token"`

	SupportsPTZ           bool    `json:"supports_ptz"`

	RetentionGBLimit      int    `json:"retention_gb_limit"`
	IsActive              bool    `json:"is_active"`
	CreatedAt             int64   `json:"created_at"`
	UpdatedAt             int64   `json:"updated_at"`
}
