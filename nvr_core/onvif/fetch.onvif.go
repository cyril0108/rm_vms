package onvif

import (
	"fmt"
	"io"
	"nvr_core/db/models"
	"regexp"

	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/media"
	xsdonvif "github.com/use-go/onvif/xsd/onvif"
)

// OnvifRecord
type OnvifRecord struct {
	IP           string  `json:"ip"`
	MACAddress   string  `json:"mac_address"`
	Manufacturer string  `json:"manufacturer"`
	Model        string  `json:"model"`
	Firmware     string  `json:"firmware"`
	SerialNumber string  `json:"serial_number"`
	SupportsPTZ  bool    `json:"supports_ptz"`

	// Streams
	MainStream       string  `json:"mainstream"`
	SubStream        string  `json:"substream"`
	MainStreamToken  string  `json:"mainstream_token"`
	SubStreamToken   string  `json:"substream_token"`

	// Camera user/pwd
	Username         string  `json:"username"`
	Password         string  `json:"password"`

	ErrorMSG         string  `json:"error_msg"`
}

func (cr *OnvifRecord) MapToDBCamera() *models.Camera {
	return &models.Camera{
		Manufacturer:          cr.Manufacturer,
		Model:                 cr.Model,
		SerialNumber:          cr.SerialNumber,
		IPAddress:             cr.IP,
		Type:                  models.CameraTypeONVIF,
		Username:              cr.Username,
		StreamURL:             cr.MainStream,
		SubStreamURL:          cr.SubStream,
		OnvifProfileToken:     cr.MainStreamToken,
		SubStreamProfileToken: cr.SubStreamToken,
		SupportsPTZ:           cr.SupportsPTZ,
	}
}

// FetchCameraONVIFData connects to an ONVIF device and extracts its DB-ready metadata
func FetchCameraONVIFData(ip string, username string, password string) (*OnvifRecord, error) {
	address := fmt.Sprintf("%s:80", ip)

	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:    address,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ONVIF device: %w", err)
	}

	record := &OnvifRecord{
		IP:       ip,
		Username: username,
		Password: password,
	}

	// Fetch Device Information (Firmware, Serial, etc.)
	devInfoReq := device.GetDeviceInformation{}
	resp, err := dev.CallMethod(devInfoReq)
	if err == nil && resp != nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		record.Manufacturer = extractTag(body, "Manufacturer")
		record.Model = extractTag(body, "Model")
		record.Firmware = extractTag(body, "FirmwareVersion")
		record.SerialNumber = extractTag(body, "SerialNumber")
	}

	// Fetch Network Interfaces (For MAC Address)
	netReq := device.GetNetworkInterfaces{}
	resp, err = dev.CallMethod(netReq)
	if err == nil && resp != nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		// ONVIF standard uses HwAddress for MAC
		record.MACAddress = extractTag(body, "HwAddress")
		if record.MACAddress == "" {
			record.MACAddress = extractTag(body, "hwAddress") // Fallback for case sensitivity
		}
	}

	// Fetch Media Profiles (To get the tokens)
	profilesReq := media.GetProfiles{}
	resp, err = dev.CallMethod(profilesReq)
	if err != nil || resp == nil {
		return record, fmt.Errorf("could not get media profiles: %w", err)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Extract ALL profile tokens as a slice
	tokens := extractTokens(body)
	if len(tokens) == 0 {
		return record, fmt.Errorf("no media profiles found on device")
	}

	// Assign Main Stream (Usually the first token)
	record.MainStreamToken = tokens[0]
	record.MainStream = fetchStreamUri(dev, record.MainStreamToken)

	// Assign Sub Stream (Usually the second token, if it exists)
	if len(tokens) > 1 {
		record.SubStreamToken = tokens[1]
		record.SubStream = fetchStreamUri(dev, record.SubStreamToken)
	}

	return record, nil
}

// fetchStreamUri is a helper to grab the RTSP URL for a specific profile token
func fetchStreamUri(dev *goonvif.Device, token string) string {
	uriReq := media.GetStreamUri{
		StreamSetup: xsdonvif.StreamSetup{
			Stream: "RTP-Unicast",
			Transport: xsdonvif.Transport{
				Protocol: "RTSP",
			},
		},
		ProfileToken: xsdonvif.ReferenceToken(token),
	}

	resp, err := dev.CallMethod(uriReq)
	if err == nil && resp != nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return extractTag(body, "Uri")
	}
	return ""
}

// extractTag is a lightweight helper to grab values from ONVIF SOAP XML, ignoring namespaces
func extractTag(xmlData []byte, tag string) string {
	re := regexp.MustCompile(`<(?:\w+:)?` + tag + `(?:[^>]*)>([^<]+)</(?:\w+:)?` + tag + `>`)
	match := re.FindSubmatch(xmlData)
	if len(match) > 1 {
		return string(match[1])
	}
	return ""
}

// extractTokens parses ALL profile tokens from the GetProfiles response
func extractTokens(xmlData []byte) []string {
	// Matches token="Profile_1", token="Profile_2", etc.
	re := regexp.MustCompile(`token="([^"]+)"`)

	// -1 means find all matches
	matches := re.FindAllSubmatch(xmlData, -1) 

	var tokens []string
	for _, match := range matches {
		if len(match) > 1 {
			tokens = append(tokens, string(match[1]))
		}
	}
	return tokens
}