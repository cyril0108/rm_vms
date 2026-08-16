package onvif

import (
	"context"
	"fmt"
	"time"

	nvrxsd "nvr_core/onvif/xml"
	nvrptz "nvr_core/onvif/xml/ptz"

	"github.com/use-go/onvif/ptz"
	xxsd "github.com/use-go/onvif/xsd"
	xsd "github.com/use-go/onvif/xsd/onvif"
)

// GetProfiles
func (m *ONVIFManager) GetPTZStatus(ctx context.Context, profile string) (nvrxsd.PTZStatus, error) {
	req := ptz.GetStatus {
		ProfileToken: xsd.ReferenceToken(profile),
	}

	data, err := CallONVIFRequest[nvrptz.GetStatusResponse](m.Device, req)
	if err != nil {
		return nvrxsd.PTZStatus{}, fmt.Errorf("failed to fetch PTZ status: %w", err)
	}

	return data.PTZStatus, nil
}

// func (m *ONVIFManager) PTZContinuousMove(ctx context.Context, profile string, velocity onvif.PTZSpeed, timeout string) (error) {
func (m *ONVIFManager) PTZContinuousMove(ctx context.Context, profile string, velocity *xsd.PTZSpeed) (error) {

	req := ptz.ContinuousMove{
		ProfileToken: xsd.ReferenceToken(profile),
		Velocity: *velocity,
		// Timeout: ,
	}

	_, err := CallONVIFRequest[nvrptz.ContinuousMoveResponse](m.Device, req)
	if err != nil {
		return fmt.Errorf("failed to fetch PTZ status: %w", err)
	}

	return nil
}

func (m *ONVIFManager) PTZRelativeMove(ctx context.Context, profile string, velocity *xsd.PTZVector, speed *xsd.PTZSpeed) (error) {

	req := ptz.RelativeMove{
		ProfileToken: xsd.ReferenceToken(profile),
		Translation: *velocity,
		Speed: *speed,
	}

	_, err := CallONVIFRequest[nvrptz.RelativeMoveResponse](m.Device, req)
	if err != nil {
		return fmt.Errorf("failed to fetch PTZ status: %w", err)
	}

	return nil
}


func (m *ONVIFManager) PTZAbsoluteMove(ctx context.Context, profile string, velocity *xsd.PTZVector, speed *xsd.PTZSpeed) (error) {

	req := ptz.AbsoluteMove{
		ProfileToken: xsd.ReferenceToken(profile),
		Position: *velocity,
		Speed: *speed,
	}

	_, err := CallONVIFRequest[nvrptz.AbsoluteMoveResponse](m.Device, req)
	if err != nil {
		return fmt.Errorf("failed to fetch PTZ status: %w", err)
	}

	return nil
}

func (m *ONVIFManager) PTZStop(ctx context.Context, profile string, stopPanTilt bool, stopZoom bool) error {

	req := ptz.Stop{
		ProfileToken: xsd.ReferenceToken(profile),
		PanTilt: xxsd.Boolean(stopPanTilt),
		Zoom: xxsd.Boolean(stopZoom),
	}

	_, err := CallONVIFRequest[nvrptz.StopResponse](m.Device, req)
	if err != nil {
		return fmt.Errorf("failed to fetch PTZ status: %w", err)
	}

	return nil
}



func (m *ONVIFManager) PTZStep(ctx context.Context, profile string, pan, tilt, zoom float64, speed float64) error {

	// TryLock instantly returns false if another Step is currently running.
	if !m.stepMu.TryLock() {
		return fmt.Errorf("PTZ step already in progress")
	}
	// Ensure we release the lock no matter how this function exits
	defer m.stepMu.Unlock()

	velocity := xsdPTZSpeed(pan, tilt, zoom)

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
	err := m.PTZContinuousMove(ctx, profile, velocity)
	if err != nil {
		return fmt.Errorf("fallback ContinuousMove failed: %w", err)
	}

	// Let the motors run for a brief moment (e.g., 300 milliseconds).
	// You can expose this duration as a configuration setting later if needed.
	time.Sleep(200 * time.Millisecond)

	// Force the motors to stop
	err = m.PTZStop(ctx, profile, true, true)
	if err != nil {
		return fmt.Errorf("failed to stop fallback step: %w", err)
	}

	return nil
}


// ===========================================================
// Private
// ===========================================================
func xsdPTZVector(pan, tilt, zoom float64) *xsd.PTZVector {
	return &xsd.PTZVector{
		PanTilt: nvrPanTiltVector(pan, tilt),
		Zoom: nvrZoomVector(zoom),
	}
}

func xsdPTZSpeed(pan, tilt, zoom float64) *xsd.PTZSpeed {
	return &xsd.PTZSpeed{
		PanTilt: nvrPanTiltVector(pan, tilt),
		Zoom: nvrZoomVector(zoom),
	}
}

// func xsdPanTiltVector(pan, tilt float64) xsd.Vector2D {
// 	return xsd.Vector2D{
// 		X:     pan,
// 		Y:     tilt,
// 		Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace",
// 		// Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace",
// 	}
// }


// func nvrZoomVector(zoom float64) xsd.Vector1D {
// 	return xsd.Vector1D{
// 		X:     zoom,
// 		Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace",
// 		// Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace",
// 	}
// }

// func nvrPTZSpeed(pan, tilt, zoom float64) xsd.PTZSpeed {
// 	return xsd.PTZSpeed{
// 		PanTilt: xsd.Vector2D{
// 			X:     pan,
// 			Y:     tilt,
// 			Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace",
// 		},
// 		Zoom: xsd.Vector1D{
// 			X:     zoom,
// 			Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace",
// 		},
// 	}
// }