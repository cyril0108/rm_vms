package onvif

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nvr_core/db/models"

	"nvr_core/onvif/xml"
	xsd "github.com/use-go/onvif/xsd/onvif"
)

// PTZPMController holds the authenticated device ready for commands
type PTZPMController struct {
	Manager      *ONVIFManager
	ProfileToken string
	Profile      *xml.Profile
	stepMu       sync.Mutex
}


// NewPTZPMController initializes a device specifically for sending PTZ commands,
func NewPTZPMController(cam *models.Camera, profileToken string, port int, key []byte) (*PTZPMController, error) {
	// address := fmt.Sprintf("%s:80", ip)

	ll := LOG.Prefix("[NewPTZPMController]")

	// ll.Info("argvs", "address", address, "username", username, "profile", profileToken)

	// address = fmt.Sprintf("http://%s/onvif/device_service", address)

	// LOG.Info("[NewPTZPMController] client address", "address", address)

	cfg, err := NewDeviceConfig(cam, port, key)
	if err != nil {
		ll.Info("Failed to create device config")
	}

	// Initialize Modern Client (Handles WS-Security & Digest Auth internally)
	manager, err := NewONVIFManager(cfg, cam)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ONVIF client: %w", err)
	}

	return &PTZPMController{
		Manager:       manager,
		ProfileToken: profileToken,
		// Profile: profile,
	}, nil
}


// MoveContinuous starts moving the camera.
// pan, tilt, and zoom values MUST be floats between -1.0 and 1.0.
// e.g., pan: 1.0 (Right), pan: -1.0 (Left), tilt: 1.0 (Up)
func (pc *PTZPMController) MoveContinuous(ctx context.Context, pan, tilt, zoom float64) error {

	velocity := &xsd.PTZSpeed{
	    PanTilt: xsd.Vector2D{X: pan, Y: tilt},
	    Zoom: xsd.Vector1D{X: zoom},
	}

	// timeout := "PT2S" // 2 seconds

	err := pc.Manager.PTZContinuousMove(ctx, pc.ProfileToken, velocity)
	if err != nil {
		return fmt.Errorf("ContinuousMove failed: %w", err)
	}

	return nil

}

// Stop halts all current movement.
// CRITICAL: If you don't call this, the camera will spin forever until it hits the physical limit!
func (pc *PTZPMController) Stop(ctx context.Context, stopPanTilt, stopZoom bool) error {

	err := pc.Manager.PTZStop(ctx, pc.ProfileToken, stopPanTilt, stopZoom)

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
func (pc *PTZPMController) GetStatus(ctx context.Context, profile string) (xml.PTZStatus, error) {
	if profile == "" {
		return pc.Manager.GetPTZStatus(ctx, pc.ProfileToken)
	}
	return pc.Manager.GetPTZStatus(ctx, profile)
}

// // Get list of PTZ presets
// func (pc *PTZPMController) GetPresets(ctx context.Context) ([]*onvif.PTZPreset, error) {
// 	return pc.Manager.PTZGetPresets(ctx, pc.ProfileToken)
// }

// // Move to a preset position
// func (pc *PTZPMController) GotoPreset(ctx context.Context, presetToken string, speed *onvif.PTZSpeed) error {
// 	return pc.Manager.PTZGotoPreset(ctx, pc.ProfileToken, presetToken, speed)
// }

// Save current position as preset
// func (pc *PTZPMController) SetPreset(ctx context.Context, ) (string, error) {
// 	return pc.Manager.SetPreset(ctx, pc.ProfileToken, )
// }

// // Delete a preset
// func (pc *PTZPMController) RemovePreset(ctx context.Context) {
// 	pc.Manager.RemovePreset(ctx, pc.ProfileToken)
// }

// Move to home position 
// func (pc *PTZPMController) GotoHomePosition(ctx context.Context, speed *onvif.PTZSpeed) error {
// 	return pc.Manager.PTZGotoHomePosition(ctx, pc.ProfileToken, speed)
// }

// // Set current position as home
// func (pc *PTZPMController) SetHomePosition(ctx context.Context) error {
// 	return pc.Manager.PTZSetHomePosition(ctx, pc.ProfileToken)
// }

// // Get PTZ configuration
// func (pc *PTZPMController) GetConfiguration(ctx context.Context) (*onvif.PTZConfiguration, error) {
// 	return pc.Manager.PTZGetConfiguration(ctx, pc.ProfileToken)
// }

// func (pc *PTZPMController) GetConfigurations(ctx context.Context) ([]*onvif.PTZConfiguration, error) {
// 	return pc.Manager.PTZGetConfigurations(ctx)
// }

// // Get all PTZ configurations
// func (pc *PTZPMController) GetCompatiblePTZConfigurations(ctx context.Context) ([]*onvif.PTZConfiguration, error) {
// 	return pc.Manager.PTZGetCompatiblePTZConfigurations(ctx, pc.ProfileToken)
// }



// MoveRelative nudges the camera a specific discrete distance.
// pan, tilt, and zoom values represent the translation step (e.g., +0.1 or -0.1).
// speed dictates how fast it makes that step (0.0 to 1.0).
func (pc *PTZPMController) MoveRelative(ctx context.Context, pan, tilt, zoom float64, speed float64) error {

	vector := xsd.PTZVector{
	    PanTilt: nvrPanTiltVector(pan, tilt),
	    Zoom:    nvrZoomVector(zoom),
	}

	velocity := nvrPTZSpeed(speed,speed,speed)

	err := pc.Manager.PTZRelativeMove(ctx, pc.ProfileToken, &vector, &velocity)
	if err != nil {
		return fmt.Errorf("RelativeMove failed: %w", err)
	}

	return nil
}

func (pc *PTZPMController) MoveAbsolute(ctx context.Context, pan, tilt, zoom float64, speed float64) error {

	position := xsd.PTZVector{
	    PanTilt: nvrPanTiltVector(pan, tilt),
	    Zoom:    nvrZoomVector(zoom),
	}

	velocity := nvrPTZSpeed(speed,speed,speed)

	err := pc.Manager.PTZAbsoluteMove(ctx, pc.ProfileToken, &position, &velocity)
	if err != nil {
		return fmt.Errorf("MoveAbsolute failed: %w", err)
	}

	return nil
}


// Step attempts a native ONVIF RelativeMove. If the camera rejects it, 
// it falls back to a time-based ContinuousMove to simulate the step.
func (pc *PTZPMController) Step(ctx context.Context, pan, tilt, zoom float64, speed float64) error {

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

func (pc *PTZPMController) MoveAbsoluteCenter(ctx context.Context) error {
	return pc.MoveAbsolute(ctx, 0, 0, 0, 0.5)
}

// ===========================================================
// Stepping Movements
// ===========================================================

func (pc *PTZPMController) StepLeft(ctx context.Context) error {
	return pc.Step(ctx, -0.1, 0, 0, 0.5)
}

func (pc *PTZPMController) StepRight(ctx context.Context) error {
	return pc.Step(ctx, 0.1, 0, 0, 0.5)
}

func (pc *PTZPMController) StepUp(ctx context.Context) error {
	return pc.Step(ctx, 0, 0.1, 0, 0.5)
}

func (pc *PTZPMController) StepDown(ctx context.Context) error {
	return pc.Step(ctx, 0, -0.1, 0, 0.5)
}

func (pc *PTZPMController) StepZoomIn(ctx context.Context) error {
	return pc.Step(ctx, 0, 0, 0.1, 0.5)
}

func (pc *PTZPMController) StepZoomOut(ctx context.Context) error {
	return pc.Step(ctx, 0, 0, -0.1, 0.5)
}

func  (pc *PTZPMController) Speed(speed float64) *xml.PTZSpeed {
	return &xml.PTZSpeed{
	    PanTilt: xml.Vector2D{X: speed, Y: speed},
	    Zoom: xml.Vector1D{X: speed},
	}
}

// ===========================================================
// Private

//
func nvrPanTiltVector(pan, tilt float64) xsd.Vector2D {
	return xsd.Vector2D{
		X:     pan,
		Y:     tilt,
		Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace",
		// Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace",
	}
}


func nvrZoomVector(zoom float64) xsd.Vector1D {
	return xsd.Vector1D{
		X:     zoom,
		Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace",
		// Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace",
	}
}

func nvrPTZSpeed(pan, tilt, zoom float64) xsd.PTZSpeed {
	return xsd.PTZSpeed{
		PanTilt: xsd.Vector2D{
			X:     pan,
			Y:     tilt,
			Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace",
		},
		Zoom: xsd.Vector1D{
			X:     zoom,
			Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace",
		},
	}
}

// ===========================================================
// Experimental Helpers


// PTZCapabilities holds the verified features of the camera
// type PTZCapabilities struct {
// 	SupportsPanTilt   bool
// 	SupportsZoom      bool
// 	SupportsRelative  bool
// 	SupportsContinuous bool
// 	SupportsAbsolute  bool
// }

func (pc *PTZPMController) PTZToken() string {
	return string(pc.Profile.PTZConfiguration.Token)
}

// ProbeCapabilities queries the camera's PTZ Configurations to see what movements it officially supports.
func (pc *PTZPMController) ProbeCapabilities(ctx context.Context) (*PTZCapabilities, error) {

	ll := LOG.Prefix("[ProbeCapabilities]")

	caps := &PTZCapabilities{}

	// Fetch the Media Profile using your "MainStream" ProfileToken
	// Note: Depending on the 0x524a/onvif-go library version, this might be called on pc.Manager.Media.GetProfile or directly on pc.Manager
	profile, err := pc.Manager.GetProfile(ctx, pc.ProfileToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get media profile: %w", err)
	}

	ll.Info("", "profile", profile)


	// Extract the actual PTZ Configuration Token mapped to this specific video stream
	if profile.PTZConfiguration.Token == "" {
		// If there is no PTZ Configuration attached to this profile, the camera definitively does not support PTZ!
		return caps, nil
	}

	// PTZToken := profile.PTZConfiguration.Token 
	// e.g., actualPTZToken is now "PTZ_Config_1" instead of "MainStream"


	// NOW fetch the PTZ Configuration using the correct token
	cfg := profile.PTZConfiguration

	ll.Info("PTZ", "token", cfg.Token)

	ll.Info("PTZConfiguration", "cft", cfg)

	// Check for Pan/Tilt support (Do the physical Pan/Tilt boundaries exist?)
	if RangeExists(cfg.PanTiltLimits.Range.XRange) || RangeExists(cfg.PanTiltLimits.Range.YRange) {
		caps.SupportsPanTilt = true
	}

	// Check for Zoom support (Do the physical Zoom boundaries exist?)
	if RangeExists(cfg.ZoomLimits.Range.XRange) {
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