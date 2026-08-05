package onvif

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/0x524a/onvif-go"

	"nvr_core/logger"
)

var LOG = logger.NewLogger("[onvif][ptz]")

// PTZController holds the authenticated device ready for commands
type PTZController struct {
	Client       *onvif.Client
	ProfileToken string
	Profile      *onvif.Profile
	stepMu       sync.Mutex
}

// PTZ Service
// https://github.com/0x524a/onvif-go#ptz-service

// NewPTZController initializes a device specifically for sending PTZ commands,
func NewPTZController(address, username, password, profileToken string) (*PTZController, error) {
	// address := fmt.Sprintf("%s:80", ip)

	ll := LOG.Prefix("[NewPTZController]")

	ll.Info("argvs", "address", address, "username", username, "profile", profileToken)

	// address = fmt.Sprintf("http://%s/onvif/device_service", address)

	// LOG.Info("[NewPTZController] client address", "address", address)

	// Initialize Modern Client (Handles WS-Security & Digest Auth internally)
	client, err := ONVIFClient(address, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ONVIF client: %w", err)
	}

	ctx := context.Background()

	if err := client.Initialize(ctx); err != nil {
		ll.Warn("failed to initialize and found ONVIF service endpoints")
	}

	profile, err := client.GetProfile(ctx, profileToken)
	if err != nil {
		ll.Info("Failed to get profile")
	}

	// info, err := client.GetDeviceInformation(ctx)
	// if err == nil {
	// 	LOG.Info("info", "device_info", info)
	// }

	// caps, err := client.GetCapabilities(ctx)
	// if err == nil {
	// 	LOG.Info("info", "Capabilities", caps)
	// }

	// endpoint, err := client.GetEndpointReference(ctx)
	// if err == nil {
	// 	LOG.Info("info", "endpoint", endpoint)
	// } else {
	// 	// LOG.Info("endpoint error", "err", err)
	// }

	// services, err := client.GetServices(ctx, true)
	// if err == nil {
	// 	LOG.Info("info", "services", services)
	// }

	// if len(services) > 0 && services[0] != nil {
	// 	LOG.Info("info", "first_service", services[0].XAddr)
	// }

	// status, err := client.GetStatus(ctx, profileToken)
	// if err == nil {
	// 	LOG.Info("info", "ptz_status", status)
	// } else {
	// 	// LOG.Error("failed to get ptz status", "err", err)
	// }

	// client.GetCompatiblePTZConfigurations(ctx, profileToken)

	// client.Endpoint()

	return &PTZController{
		Client:       client,
		ProfileToken: profileToken,
		Profile: profile,
	}, nil
}


// MoveContinuous starts moving the camera.
// pan, tilt, and zoom values MUST be floats between -1.0 and 1.0.
// e.g., pan: 1.0 (Right), pan: -1.0 (Left), tilt: 1.0 (Up)
func (pc *PTZController) MoveContinuous(ctx context.Context, pan, tilt, zoom float64) error {

	velocity := &onvif.PTZSpeed{
	    PanTilt: &onvif.Vector2D{X: pan, Y: tilt},
	    Zoom: &onvif.Vector1D{X: zoom},
	}

	timeout := "PT2S" // 2 seconds

	err := pc.Client.ContinuousMove(ctx, pc.ProfileToken, velocity, &timeout)
	if err != nil {
		return fmt.Errorf("ContinuousMove failed: %w", err)
	}

	return nil

}

// Stop halts all current movement.
// CRITICAL: If you don't call this, the camera will spin forever until it hits the physical limit!
func (pc *PTZController) Stop(ctx context.Context, stopPanTilt, stopZoom bool) error {

	err := pc.Client.Stop(ctx, pc.ProfileToken, stopPanTilt, stopZoom)

	if err != nil {
		return fmt.Errorf("Stop failed: %w", err)
	}

	return nil
}

// =============
// PTZ Service
// =============

// ContinuousMove()
// Start continuous PTZ movement
// AbsoluteMove()
// Move to absolute position
// RelativeMove()
// Move relative to current position

// Stop()
// Stop PTZ movement

// Get current PTZ status and position
func (pc *PTZController) GetStatus(ctx context.Context, profile string) (*onvif.PTZStatus, error) {
	if profile == "" {
		return pc.Client.GetStatus(ctx, pc.ProfileToken)
	}
	return pc.Client.GetStatus(ctx, profile)
}

// Get list of PTZ presets
func (pc *PTZController) GetPresets(ctx context.Context) ([]*onvif.PTZPreset, error) {
	return pc.Client.GetPresets(ctx, pc.ProfileToken)
}

// Move to a preset position
func (pc *PTZController) GotoPreset(ctx context.Context, presetToken string, speed *onvif.PTZSpeed) error {
	return pc.Client.GotoPreset(ctx, pc.ProfileToken, presetToken, speed)
}

// Save current position as preset
// func (pc *PTZController) SetPreset(ctx context.Context, ) (string, error) {
// 	return pc.Client.SetPreset(ctx, pc.ProfileToken, )
// }

// // Delete a preset
// func (pc *PTZController) RemovePreset(ctx context.Context) {
// 	pc.Client.RemovePreset(ctx, pc.ProfileToken)
// }

// Move to home position 
func (pc *PTZController) GotoHomePosition(ctx context.Context, speed *onvif.PTZSpeed) error {
	return pc.Client.GotoHomePosition(ctx, pc.ProfileToken, speed)
}

// Set current position as home
func (pc *PTZController) SetHomePosition(ctx context.Context) error {
	return pc.Client.SetHomePosition(ctx, pc.ProfileToken)
}

// Get PTZ configuration
func (pc *PTZController) GetConfiguration(ctx context.Context) (*onvif.PTZConfiguration, error) {
	return pc.Client.GetConfiguration(ctx, pc.ProfileToken)
}

func (pc *PTZController) GetConfigurations(ctx context.Context) ([]*onvif.PTZConfiguration, error) {
	return pc.Client.GetConfigurations(ctx)
}

// Get all PTZ configurations
func (pc *PTZController) GetCompatiblePTZConfigurations(ctx context.Context) ([]*onvif.PTZConfiguration, error) {
	return pc.Client.GetCompatiblePTZConfigurations(ctx, pc.ProfileToken)
}



// MoveRelative nudges the camera a specific discrete distance.
// pan, tilt, and zoom values represent the translation step (e.g., +0.1 or -0.1).
// speed dictates how fast it makes that step (0.0 to 1.0).
func (pc *PTZController) MoveRelative(ctx context.Context, pan, tilt, zoom float64, speed float64) error {

	vector := &onvif.PTZVector{
	    PanTilt: panTiltVector(pan, tilt),
	    Zoom:    zoomVector(zoom),
	}

	velocity := ptzSpeed(speed,speed,speed)

	err := pc.Client.RelativeMove(ctx, pc.ProfileToken, vector, velocity)
	if err != nil {
		return fmt.Errorf("RelativeMove failed: %w", err)
	}

	return nil
}

func (pc *PTZController) MoveAbsolute(ctx context.Context, pan, tilt, zoom float64, speed float64) error {

	position := &onvif.PTZVector{
	    PanTilt: panTiltVector(pan, tilt),
	    Zoom:    zoomVector(zoom),
	}

	velocity := ptzSpeed(speed,speed,speed)

	err := pc.Client.AbsoluteMove(ctx, pc.ProfileToken, position, velocity)
	if err != nil {
		return fmt.Errorf("MoveAbsolute failed: %w", err)
	}

	return nil
}


// Step attempts a native ONVIF RelativeMove. If the camera rejects it, 
// it falls back to a time-based ContinuousMove to simulate the step.
func (pc *PTZController) Step(ctx context.Context, pan, tilt, zoom float64, speed float64) error {

	// TryLock instantly returns false if another Step is currently running.
	if !pc.stepMu.TryLock() {
		return fmt.Errorf("PTZ step already in progress")
	}
	// Ensure we release the lock no matter how this function exits
	defer pc.stepMu.Unlock()

	// Try the mathematically correct ONVIF Relative Move
	// err := pc.MoveRelative(ctx, pan, tilt, zoom, speed)

	// if err == nil {
	// 	// The camera supported it and executed perfectly!
	// 	return nil
	// }

	// FALLBACK: The camera rejected RelativeMove (likely a generic camera).
	// We simulate a step using ContinuousMove + Sleep + Stop.

	// relErr := err

	// Start the motors
	err := pc.MoveContinuous(ctx, pan, tilt, zoom)
	if err != nil {
		return fmt.Errorf("fallback ContinuousMove failed: %w", err)
	}

	// Let the motors run for a brief moment (e.g., 300 milliseconds).
	// You can expose this duration as a configuration setting later if needed.
	time.Sleep(1000 * time.Millisecond)

	// Force the motors to stop
	err = pc.Stop(ctx, true, true)
	if err != nil {
		return fmt.Errorf("failed to stop fallback step: %w", err)
	}

	return nil
}


// ===========================================================
// Absolute Movements
// ===========================================================

func (pc *PTZController) MoveAbsoluteCenter(ctx context.Context) error {
	return pc.MoveAbsolute(ctx, 0, 0, 0, 0.5)
}

// ===========================================================
// Stepping Movements
// ===========================================================

func (pc *PTZController) StepLeft(ctx context.Context) error {
	return pc.Step(ctx, -0.1, 0, 0, 0.5)
}

func (pc *PTZController) StepRight(ctx context.Context) error {
	return pc.Step(ctx, 0.1, 0, 0, 0.5)
}

func (pc *PTZController) StepUp(ctx context.Context) error {
	return pc.Step(ctx, 0, 0.1, 0, 0.5)
}

func (pc *PTZController) StepDown(ctx context.Context) error {
	return pc.Step(ctx, 0, -0.1, 0, 0.5)
}

func (pc *PTZController) StepZoomIn(ctx context.Context) error {
	return pc.Step(ctx, 0, 0, 0.1, 0.5)
}

func (pc *PTZController) StepZoomOut(ctx context.Context) error {
	return pc.Step(ctx, 0, 0, -0.1, 0.5)
}

func  (pc *PTZController) Speed(speed float64) *onvif.PTZSpeed {
	return &onvif.PTZSpeed{
	    PanTilt: &onvif.Vector2D{X: speed, Y: speed},
	    Zoom: &onvif.Vector1D{X: speed},
	}
}

// ===========================================================
// Private
func panTiltVector(pan, tilt float64) *onvif.Vector2D {
	return &onvif.Vector2D{
		X:     pan,
		Y:     tilt,
		Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace",
		// Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace",
	}
}


func zoomVector(zoom float64) *onvif.Vector1D {
	return &onvif.Vector1D{
		X:     zoom,
		Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace",
		// Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace",
	}
}

func ptzSpeed(pan, tilt, zoom float64) *onvif.PTZSpeed {
	return &onvif.PTZSpeed{
		PanTilt: &onvif.Vector2D{
			X:     pan,
			Y:     tilt,
			Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace",
		},
		Zoom: &onvif.Vector1D{
			X:     zoom,
			Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace",
		},
	}
}

// ===========================================================
// Experimental Helpers


// PTZCapabilities holds the verified features of the camera
type PTZCapabilities struct {
	SupportsPanTilt   bool
	SupportsZoom      bool
	SupportsRelative  bool
	SupportsContinuous bool
	SupportsAbsolute  bool
}

func (pc *PTZController) PTZToken() string {
	if pc.Profile != nil && pc.Profile.PTZConfiguration != nil {
		return pc.Profile.PTZConfiguration.Token
	}
	return ""
}

// ProbeCapabilities queries the camera's PTZ Configurations to see what movements it officially supports.
func (pc *PTZController) ProbeCapabilities(ctx context.Context) (*PTZCapabilities, error) {

	ll := LOG.Prefix("[ProbeCapabilities]")

	caps := &PTZCapabilities{}

	// 1. Fetch the Media Profile using your "MainStream" ProfileToken
	// Note: Depending on the 0x524a/onvif-go library version, this might be called on pc.Client.Media.GetProfile or directly on pc.Client
	profile, err := pc.Client.GetProfile(ctx, pc.ProfileToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get media profile: %w", err)
	}

	ll.Info("", "profile", profile)


	// 2. Extract the actual PTZ Configuration Token mapped to this specific video stream
	if profile.PTZConfiguration == nil || profile.PTZConfiguration.Token == "" {
		// If there is no PTZ Configuration attached to this profile, the camera definitively does not support PTZ!
		return caps, nil
	}

	// PTZToken := profile.PTZConfiguration.Token 
	// e.g., actualPTZToken is now "PTZ_Config_1" instead of "MainStream"


	// 3. NOW fetch the PTZ Configuration using the correct token
	cfg := profile.PTZConfiguration

	ll.Info("PTZ", "token", cfg.Token)

	ll.Info("PTZConfiguration", "cft", cfg)

	// Check for Pan/Tilt support (Do the physical Pan/Tilt boundaries exist?)
	if cfg.PanTiltLimits != nil {
		caps.SupportsPanTilt = true
	}
	
	// Check for Zoom support (Do the physical Zoom boundaries exist?)
	if cfg.ZoomLimits != nil {
		caps.SupportsZoom = true
	}

	// -------------------------------------------------------------------------
	// STRUCT WARNING: Depending on how the library generated the WSDL structs,
	// these space properties might be pointers (*string) or direct strings (string).
	// If your Go compiler complains about `!= nil` below, simply change it to `!= ""`
	// -------------------------------------------------------------------------

	// Check for Continuous Move support
	if cfg.DefaultContinuousPanTiltVelocitySpace != "" || cfg.DefaultContinuousZoomVelocitySpace != "" {
		caps.SupportsContinuous = true
	}

	// Check for Relative Move support (Stepping)
	if cfg.DefaultRelativePanTiltTranslationSpace != "" || cfg.DefaultRelativeZoomTranslationSpace != "" {
		caps.SupportsRelative = true
	}
	
	// Check for Absolute Move support
	if cfg.DefaultAbsolutePantTiltPositionSpace != "" || cfg.DefaultAbsoluteZoomPositionSpace != "" {
		caps.SupportsAbsolute = true
	}

	return caps, nil
}