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

// OnvifRecord represents the exact structure you will save to SQLite
type OnvifRecord struct {
	IP           string
	Manufacturer string
	Model        string
	Firmware     string
	SerialNumber string
	SupportsPTZ  bool
	// Streams
	MainStream       string
	SubStream        string
	MainStreamToken  string
	SubStreamToken   string
	// Camera user/pwd
	Username         string
	Password         string
}

func (cr *OnvifRecord) MapToDBCamera() *models.Camera {
	// ID should be created elsewhere.
	return &models.Camera{
		Manufacturer:     cr.Manufacturer,
		Model:            cr.Model,
		SerialNumber:     cr.SerialNumber,
		IPAddress:        cr.IP,
		Type:             models.CameraTypeONVIF,
		Username:         cr.Username,
		StreamURL:        cr.MainStream,
		SubStreamURL:     cr.SubStream,
		OnvifProfileToken: cr.MainStreamToken,
		SubStreamProfileToken: cr.SubStreamToken,
		SupportsPTZ:      cr.SupportsPTZ,
		// RetentionGBLimit: cr.RetentionGBLimit,
		// IsActive:         cr.IsActive,
	}
}


// FetchCameraONVIFData connects to an ONVIF device and extracts its DB-ready metadata
func FetchCameraONVIFData(ip string, username string, password string) (*OnvifRecord, error) {
	address := fmt.Sprintf("%s:80", ip)

	// Initialize authenticated ONVIF device client
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

	// Fetch Device Information
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

	// Fetch Media Profiles
	profilesReq := media.GetProfiles{}
	resp, err = dev.CallMethod(profilesReq)
	if err != nil || resp == nil {
		return record, fmt.Errorf("could not get media profiles: %w", err)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Extract the first profile token (usually the main stream)
	mainProfileToken := extractToken(body)
	if mainProfileToken == "" {
		return record, fmt.Errorf("no media profiles found on device")
	}

	// Fetch the RTSP Stream URI
	uriReq := media.GetStreamUri{
		// Note the xsdonvif package usage here
		StreamSetup: xsdonvif.StreamSetup{
			Stream: "RTP-Unicast",
			Transport: xsdonvif.Transport{
				Protocol: "RTSP",
			},
		},
		ProfileToken: xsdonvif.ReferenceToken(mainProfileToken),
	}

	resp, err = dev.CallMethod(uriReq)
	if err == nil && resp != nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		record.MainStream = extractTag(body, "Uri")
	}

	return record, nil
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

// extractToken parses the first profile token from the GetProfiles response
func extractToken(xmlData []byte) string {
	// Matches token="Profile_1" or similar
	re := regexp.MustCompile(`token="([^"]+)"`)
	match := re.FindSubmatch(xmlData)
	if len(match) > 1 {
		return string(match[1])
	}
	return ""
}