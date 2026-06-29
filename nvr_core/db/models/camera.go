package models

import (
	"net/url"
	"strconv"

	"nvr_core/network"
	"nvr_core/security"
)

const CameraTypeONVIF = "onvif"
const CameraTypeRTSP  = "rtsp"

// Camera represents an IP camera (ONVIF or generic RTSP)
type Camera struct {
	ID                    int64   `json:"id"`
	UUID                  string  `json:"uuid"`
	Name                  string  `json:"name"`

	Manufacturer          string `json:"manufacturer"`
	Model                 string `json:"model"`
	SerialNumber          string  `json:"serial_number"`

	MACAddress            string  `json:"mac_address"`

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

func (c *Camera) DefaultName() string {
	name := ""
	if c.Model != "" {
		name = name + c.Model
	}
	// 
	if name == "" {
		name = c.IPAddress
	} else {
		_, i, err := network.GetSubnetParts(c.IPAddress)
		if err != nil {
			name = name + strconv.Itoa(i)
		}
	}
	return name
}

func (c *Camera) EncryptPassword(plaintext string, masterKey []byte) {

	c.PasswordEnc, _ = security.Encrypt(plaintext, masterKey)

}

func (c *Camera) DecryptPassword(masterKey []byte) (string, error) {
	return security.Decrypt(c.PasswordEnc, masterKey)
}

// Return Auth injected main stream url
func (c *Camera) AuthMainUrl(masterKey []byte) (string) {
	url, err := c.injectAuth(c.StreamURL, masterKey)
	if err != nil {
		return c.StreamURL
	}
	return url
}

func (c *Camera) AuthSubUrl(masterKey []byte) (string) {
	url, err := c.injectAuth(c.SubStreamURL, masterKey)
	if err != nil {
		return c.StreamURL
	}
	return url
}

func (c *Camera) injectAuth(url string, masterKey []byte) (string, error) {

	pwd, err := c.DecryptPassword(masterKey)
	if err != nil {
		return url, err
	}

	rurl, err := injectCredentials(url, c.Username, pwd)

	if err != nil {
		return url, err
	}

	return rurl, nil

}

func injectCredentials(rawURL, username, password string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Safely injects and URL-encodes the credentials
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String(), nil
}