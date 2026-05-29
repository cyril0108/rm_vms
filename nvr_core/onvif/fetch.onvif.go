package onvif

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"

	"regexp"

	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/media"
	xsdonvif "github.com/use-go/onvif/xsd/onvif"

	"nvr_core/db/models"
	"nvr_core/utils"
)


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

	customClient := &http.Client{
		Transport: &AuthInterceptor{
			Proxied:  http.DefaultTransport,
			Username: username,
			Password: password,
		},
	}

	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:    address,
		Username: username,
		Password: password,
		HttpClient: customClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ONVIF device: %w", err)
	}

	// dev.Authenticate(username, password)

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

	// If the string "PTZConfiguration" exists anywhere in the profiles XML, 
	// the camera physically supports PTZ movement.
	// It's "Configuration", meaing you've config for it
	record.SupportsPTZ = bytes.Contains(body, []byte("PTZConfiguration"))

	// Extract ALL profile tokens as a slice
	tokens := extractTokens(body)
	if len(tokens) == 0 {
		utils.PrintSimplifiedXML(body)
		sec, err := CheckCameraTimeDrift(ip)
		if err == nil {
			return record, fmt.Errorf("no media profiles found on device. Camera time drift: %f", sec)
		}
		return record, fmt.Errorf("no media profiles found on device")
	}

	log.Printf("Found media tokens: %v", tokens)

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
		return html.UnescapeString(extractTag(body, "Uri"))
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
// extractTokens parses ONLY the main Profile tokens from the GetProfiles response,
// actively ignoring sub-configuration tokens (VideoSource, Audio, etc.)
func extractTokens(xmlData []byte) []string {
	// This regex looks specifically for the <Profiles> tag (with or without a namespace like trt:)
	// and extracts ONLY the token attribute attached directly to it.
	re := regexp.MustCompile(`<(?:\w+:)?Profiles[^>]*\btoken="([^"]+)"`)
	
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

// VerifyCredentials performs a lightweight check against the camera's locked
// media profiles to definitively prove if the username/password are correct.
func VerifyCredentials(ip, username, password string) (bool, error) {
	address := fmt.Sprintf("%s:80", ip)

	// Reuse our bulletproof interceptor to guarantee WS-Security and Basic Auth are sent
	customClient := &http.Client{
		Transport: &AuthInterceptor{
			Proxied:  http.DefaultTransport,
			Username: username,
			Password: password,
		},
	}

	// Initialize the connection
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:      address,
		Username:   username,
		Password:   password,
		HttpClient: customClient,
	})
	if err != nil {
		// Network error (camera offline, blocked port, etc.)
		return false, fmt.Errorf("network error or ONVIF not supported: %w", err)
	}

	// The Authentication Test
	// We ask for the profiles. If the credentials are bad, this throws a NotAuthorized error.
	req := media.GetProfiles{}
	resp, err := dev.CallMethod(req)
	
	if err != nil {
		// Credentials failed!
		return false, fmt.Errorf("authentication rejected by camera: %w", err)
	}

	// Clean up the HTTP response
	if resp != nil {
		resp.Body.Close()
	}

	// Success! The credentials are valid.
	return true, nil
}