package onvif

import (
	"context"
	"fmt"
	"time"

	"github.com/0x524a/onvif-go"

	"nvr_core/db/models"
	"nvr_core/logger"
)

// MapToDBCamera remains unchanged
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
		SupportsAudio:         cr.SupportsAudio,
		SupportsAudioOutput:   cr.SupportsAudioOutput,
		EnableAudio:           cr.SupportsAudio,
	}
}

func ONVIFAddress(ip string, port int) string {
	return fmt.Sprintf("%s:%d", ip, port)
}

func ONVIFClient(address, username, password string) (*onvif.Client, error) {
	return onvif.NewClient(
		address,
		onvif.WithCredentials(username, password),
	)
}

// FetchCameraONVIFData connects to an ONVIF device and extracts its DB-ready metadata
func FetchCameraONVIFData(ip string, port int, username, password string) (*OnvifRecord, error) {
	address := ONVIFAddress(ip, port)

	record := &OnvifRecord{
		IP:       ip,
		Username: username,
		Password: password,
	}

	ll := logger.NewLogger("[]")
	// ll.Debug("[AddONVIFCamera] receive", "u", username, "p", password)

	// Establish Context (CRITICAL for NVRs so Goroutines don't hang if camera dies)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Initialize Modern Client (Handles WS-Security & Digest Auth internally)
	client, err := ONVIFClient(address, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ONVIF client: %w", err)
	}

	// Fetch Device Information (Strictly Typed)
	devInfo, err := client.GetDeviceInformation(ctx)
	if err == nil && devInfo != nil {
		record.Manufacturer = devInfo.Manufacturer
		record.Model = devInfo.Model
		record.Firmware = devInfo.FirmwareVersion
		record.SerialNumber = devInfo.SerialNumber
	}

	// Fetch Network Interfaces (For MAC Address)
	// Note: Cameras can have multiple NICs, we generally grab the first one.
	netInterfaces, err := client.GetNetworkInterfaces(ctx)
	if err == nil && len(netInterfaces) > 0 {
		record.MACAddress = netInterfaces[0].Info.HwAddress 
	}

	// Fetch Media Profiles
	profiles, err := client.GetProfiles(ctx)
	if err != nil || len(profiles) == 0 {
		sec, driftErr := CheckCameraTimeDrift(ip) // Assuming this exists elsewhere in your package
		if driftErr == nil {
			return record, fmt.Errorf("no media profiles found on device. Camera time drift: %f. err: %v", sec, err)
		}
		return record, fmt.Errorf("no media profiles found on device: %w", err)
	}

	ll.Debug("[AddONVIFCamera] Found media profiles", "count", len(profiles))

	// Assign Main Stream (First Profile)
	mainProfile := profiles[0]
	record.MainStreamToken = string(mainProfile.Token)

	// Strongly typed check for PTZ capability instead of raw string searching
	if mainProfile.PTZConfiguration != nil {
		record.SupportsPTZ = true
	}


	// ==========================================
	// AUDIO CAPABILITY DETECTION
	// ==========================================
	
	// Check for standard Audio (Microphone / One-Way)
	// If the profile has an audio source or encoder, the camera can send audio to the NVR.
	if mainProfile.AudioSourceConfiguration != nil || mainProfile.AudioEncoderConfiguration != nil {
		record.SupportsAudio = true
	}

	// Many cameras physically have a speaker, but don't bind it to the default video profile.
	// We explicitly ask the ONVIF Media service if any physical audio outputs exist.
	audioOutputs, err := client.GetAudioOutputs(ctx)
	if err == nil && len(audioOutputs) > 0 {
		record.SupportsAudioOutput = true
	}
	// ==========================================




	// Fetch Stream URI for Main Stream
	if stream, err := client.GetStreamURI(ctx, record.MainStreamToken); err == nil && stream != nil {
		record.MainStream = string(stream.URI)
	}

	// Assign Sub Stream (Second Profile, if it exists)
	if len(profiles) > 1 {
		subProfile := profiles[1]
		record.SubStreamToken = string(subProfile.Token)
		
		if stream, err := client.GetStreamURI(ctx, record.SubStreamToken); err == nil && stream != nil {
			record.SubStream = string(stream.URI)
		}
	}

	return record, nil
}

// VerifyCredentials performs a lightweight check against the camera's locked
// media profiles to definitively prove if the username/password are correct.
func VerifyCredentials(ip, username, password string) (bool, error) {
	address := fmt.Sprintf("%s:80", ip)

	// Short timeout just for credential checking
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := onvif.NewClient(
		address,
		onvif.WithCredentials(username, password),
	)
	if err != nil {
		// Network error (camera offline, blocked port, etc.)
		return false, fmt.Errorf("network error or ONVIF not supported: %w", err)
	}

	// The Authentication Test
	// In typed libraries, WS-Security faults, 401s, and 403s are caught and returned as actual Go errors
	_, err = client.GetProfiles(ctx)
	if err != nil {
		// Credentials failed, or the user lacks privileges for the Media service
		return false, nil 
	}

	// Success! No messy body scraping required.
	return true, nil
}