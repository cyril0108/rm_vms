package dto

import "nvr_core/db/models"

type CreateCameraRequest struct {
	Name             string  `json:"name" validate:"required"`
	Manufacturer     string  `json:"manufacturer"` // Pointers allow JSON "null"
	Model            string  `json:"model"`
	SerialNumber     string  `json:"serial_number"`

	IPAddress        string  `json:"ip_address" validate:"required,ip"`
	HTTPPort         int     `json:"http_port"`
	Type             string  `json:"type" validate:"required,oneof=onvif rtsp"`

	Username         string `json:"username"`
	Password         string `json:"password"` // Raw password from the UI form!

	StreamURL        string  `json:"stream_url"`
	SubStreamURL     string `json:"sub_stream_url"`

	SupportsPTZ      bool    `json:"supports_ptz"`

	RetentionGBLimit int    `json:"retention_gb_limit"`
	IsActive         bool    `json:"is_active"`
}


// CameraDetailResponse is the safe, complete payload sent to the Vue frontend
type CameraDetailResponse struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Manufacturer     string  `json:"manufacturer"`
	Model            string  `json:"model"`
	SerialNumber     string  `json:"serial_number"`

	IPAddress        string  `json:"ip_address"`
	HTTPPort         int     `json:"http_port"`
	Type             string  `json:"type"`

	Username         string  `json:"username"`
	// Notice: No password field of any kind is sent back!

	StreamURL        string  `json:"stream_url"`
	SubStreamURL     string  `json:"sub_stream_url"`

	SupportsPTZ      bool    `json:"supports_ptz"`
	RetentionGBLimit int     `json:"retention_gb_limit"`
	IsActive         bool    `json:"is_active"`

	CreatedAt        int64   `json:"created_at"`
	UpdatedAt        int64   `json:"updated_at"`
}

func (cr *CreateCameraRequest) MapToDBCamera() *models.Camera {
	// ID should be created elsewhere.
	return &models.Camera{
		Name:             cr.Name,
		Manufacturer:     cr.Manufacturer,
		Model:            cr.Model,
		SerialNumber:     cr.SerialNumber,
		IPAddress:        cr.IPAddress,
		HTTPPort:         cr.HTTPPort,
		Type:             cr.Type,
		Username:         cr.Username,
		StreamURL:        cr.StreamURL,
		SubStreamURL:     cr.SubStreamURL,
		SupportsPTZ:      cr.SupportsPTZ,
		RetentionGBLimit: cr.RetentionGBLimit,
		IsActive:         cr.IsActive,
	}
}

func MapCameraToDetail(cam models.Camera) CameraDetailResponse {
	return CameraDetailResponse{
		ID:               cam.ID,
		Name:             cam.Name,
		Manufacturer:     cam.Manufacturer,
		Model:            cam.Model,
		SerialNumber:     cam.SerialNumber,
		IPAddress:        cam.IPAddress,
		HTTPPort:         cam.HTTPPort,
		Type:             cam.Type,
		Username:         cam.Username,
		StreamURL:        cam.StreamURL,
		SubStreamURL:     cam.SubStreamURL,
		SupportsPTZ:      cam.SupportsPTZ,
		RetentionGBLimit: cam.RetentionGBLimit,
		IsActive:         cam.IsActive,
		CreatedAt:        cam.CreatedAt,
		UpdatedAt:        cam.UpdatedAt,
	}
}