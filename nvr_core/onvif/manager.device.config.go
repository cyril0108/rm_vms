package onvif

import (
	"net/http"

	"nvr_core/db/models"
)

// Config holds initialization parameters
type DeviceConfig struct {
	IP       string
	Port     int
	Username string
	Password string
	// HTTPClient allows us to inject the XML Logging Interceptor
	HTTPClient *http.Client
}

func NewDeviceConfig(cam *models.Camera, port int, key []byte) (DeviceConfig, error) {

	pwd, err := cam.DecryptPassword(key)
	if err != nil {
		return DeviceConfig{}, err
	}

	if port == 0 {
		port = cam.HTTPPort
	}

	return DeviceConfig {
		IP: cam.IPAddress,
		Port: port,
		Username: cam.Username,
		Password: pwd,
	}, nil

}