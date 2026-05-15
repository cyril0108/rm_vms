package dto

import (
	"nvr_core/security"
)

// UpdateCameraRequest uses pointers so we can differentiate between
// "missing from JSON" (nil) and "explicitly set to empty/zero" (non-nil pointer).
type UpdateCameraRequest struct {
	Name                  *string `json:"name"`
	UUID                  *string `json:"uuid"`
	Manufacturer          *string `json:"manufacturer"`
	Model                 *string `json:"model"`
	SerialNumber          *string `json:"serial_number"`
	IPAddress             *string `json:"ip_address"`
	HTTPPort              *int    `json:"http_port"`
	Type                  *string `json:"type"`
	MACAddress            *string `json:"mac_address"`
	Username              *string `json:"username"`
	Password              *string `json:"password"`
	StreamURL             *string `json:"stream_url"`
	SubStreamURL          *string `json:"sub_stream_url"`
	OnvifProfileToken     *string `json:"onvif_profile_token"`
	SubStreamProfileToken *string `json:"sub_stream_profile_token"`
	SupportsPTZ           *bool   `json:"supports_ptz"`
	RetentionGBLimit      *int    `json:"retention_gb_limit"`
	IsActive              *bool   `json:"is_active"`
}

func (u *UpdateCameraRequest) ToMapInterface(masterKey []byte) map[string]interface{} {

	updates := make(map[string]interface{})

	if u.Name != nil { updates["name"] = *u.Name }
	if u.UUID != nil { updates["uuid"] = *u.UUID }
	if u.Manufacturer != nil { updates["manufacturer"] = *u.Manufacturer }
	if u.Model != nil { updates["model"] = *u.Model }
	if u.SerialNumber != nil { updates["serial_number"] = *u.SerialNumber }
	if u.IPAddress != nil { updates["ip_address"] = *u.IPAddress }
	if u.HTTPPort != nil { updates["http_port"] = *u.HTTPPort }
	if u.Type != nil { updates["type"] = *u.Type }
	if u.MACAddress != nil { updates["mac_address"] = *u.MACAddress }
	if u.Username != nil { updates["username"] = *u.Username }
	if u.StreamURL != nil { updates["stream_url"] = *u.StreamURL }
	if u.SubStreamURL != nil { updates["sub_stream_url"] = *u.SubStreamURL }
	if u.OnvifProfileToken != nil { updates["onvif_profile_token"] = *u.OnvifProfileToken }
	if u.SubStreamProfileToken != nil { updates["sub_stream_profile_token"] = *u.SubStreamProfileToken }
	if u.RetentionGBLimit != nil { updates["retention_gb_limit"] = *u.RetentionGBLimit }

	if u.Password != nil {
		if pwd, err := security.Encrypt(*u.Password, masterKey); err!=nil {
			updates["password_enc"] = pwd
		}
	}

	// Booleans require special handling because SQLite stores them as integers
	if u.SupportsPTZ != nil {
		if *u.SupportsPTZ {
			updates["supports_ptz"] = 1
		} else {
			updates["supports_ptz"] = 0
		}
	}
	if u.IsActive != nil {
		if *u.IsActive {
			updates["is_active"] = 1
		} else {
			updates["is_active"] = 0
		}
	}

	return updates

}