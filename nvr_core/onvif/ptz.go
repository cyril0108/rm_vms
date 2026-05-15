package onvif

import (
	"fmt"
	"net/http"

	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/ptz"
	"github.com/use-go/onvif/xsd"
	xsdonvif "github.com/use-go/onvif/xsd/onvif"
)

// PTZController holds the authenticated device ready for commands
type PTZController struct {
	Device       *goonvif.Device
	ProfileToken string
}

// NewPTZController initializes a device specifically for sending PTZ commands,
// reusing our bulletproof AuthInterceptor.
func NewPTZController(ip, username, password, profileToken string) (*PTZController, error) {
	address := fmt.Sprintf("%s:80", ip)

	customClient := &http.Client{
		Transport: &AuthInterceptor{
			Proxied:  http.DefaultTransport,
			Username: username,
			Password: password,
		},
	}

	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:      address,
		Username:   username,
		Password:   password,
		HttpClient: customClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect for PTZ: %w", err)
	}

	return &PTZController{
		Device:       dev,
		ProfileToken: profileToken,
	}, nil
}

// MoveContinuous starts moving the camera.
// pan, tilt, and zoom values MUST be floats between -1.0 and 1.0.
// e.g., pan: 1.0 (Right), pan: -1.0 (Left), tilt: 1.0 (Up)
func (pc *PTZController) MoveContinuous(pan, tilt, zoom float64) error {
	req := ptz.ContinuousMove{
		ProfileToken: xsdonvif.ReferenceToken(pc.ProfileToken),
		Velocity: xsdonvif.PTZSpeed{
			PanTilt: xsdonvif.Vector2D{
				X:     pan,
				Y:     tilt,
				// Hardcoded 2D Space to Generic
				Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace",
			},
			Zoom: xsdonvif.Vector1D{
				X:     zoom,
				// Hardcoded 2D Space to Generic
				Space: "http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace",
			},
		},
	}

	resp, err := pc.Device.CallMethod(req)
	if err != nil {
		return fmt.Errorf("ContinuousMove failed: %w", err)
	}
	if resp != nil {
		resp.Body.Close()
	}

	return nil
}

// Stop halts all current movement.
// CRITICAL: If you don't call this, the camera will spin forever until it hits the physical limit!
func (pc *PTZController) Stop(stopPanTilt, stopZoom bool) error {
	req := ptz.Stop{
		ProfileToken: xsdonvif.ReferenceToken(pc.ProfileToken),
		PanTilt:      xsd.Boolean(stopPanTilt),
		Zoom:         xsd.Boolean(stopZoom),
	}

	resp, err := pc.Device.CallMethod(req)
	if err != nil {
		return fmt.Errorf("Stop failed: %w", err)
	}
	if resp != nil {
		resp.Body.Close()
	}

	return nil
}