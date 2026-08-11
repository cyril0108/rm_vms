package onvif

import (
	"context"
	"fmt"

	"github.com/use-go/onvif/device"
	sdk "github.com/use-go/onvif/sdk/device"
	xsd "github.com/use-go/onvif/xsd/onvif"
)


// GetNetworkMAC fetches the physical MAC address of the camera.
// Many cheap cameras do not report a Serial Number, so the MAC address 
// becomes your only reliable hardware fingerprint for the database.
func (m *ONVIFManager) GetNetworkMAC(ctx context.Context) (string, error) {

	req := device.GetNetworkInterfaces{}

	res, err := sdk.Call_GetNetworkInterfaces(ctx, m.Device, req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch network interfaces: %w", err)
	}

	// Assuming the first interface (usually eth0) is the active one
	mac := string(res.NetworkInterfaces.Info.HwAddress)
	return mac, nil
}

// GetCapabilities asks the camera for its comprehensive XML capability tree.
// This tells you exactly what the camera supports (e.g., specific PTZ limits, Audio specs).
func (m *ONVIFManager) GetCapabilities(ctx context.Context) (*xsd.Capabilities, error) {
	req := device.GetCapabilities{
		// Ask the camera to return everything (All, Analytics, Device, Events, Imaging, Media, PTZ)
		Category: "All", 
	}

	res, err := sdk.Call_GetCapabilities(ctx, m.Device, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch detailed capabilities: %w", err)
	}

	return &res.Capabilities, nil
}
