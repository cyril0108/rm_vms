package onvif

import (
	"context"
	"fmt"

	nvrxsd   "nvr_core/onvif/xml"
	nvrmedia "nvr_core/onvif/media"

	"github.com/use-go/onvif/media"
	// sdk "github.com/use-go/onvif/sdk/media"
	xsd "github.com/use-go/onvif/xsd/onvif"
)



// GetProfiles
func (m *ONVIFManager) GetProfiles(ctx context.Context) ([]nvrxsd.Profile, error) {
	req := media.GetProfiles {}

	data, err := CallONVIFRequest[nvrmedia.GetProfilesResponse](m.Device, req)
	if err != nil {
		return []nvrxsd.Profile{}, fmt.Errorf("failed to fetch video source: %w", err)
	}

	return data.Profiles, nil
}

// This won't work. GetVideoSource does not ask for profile token.
func (m *ONVIFManager) GetMainVideoSource(ctx context.Context) (nvrxsd.VideoSourceConfiguration, error) {
	return m.GetVideoSource(ctx, m.Camera.OnvifProfileToken)
}

func (m *ONVIFManager) GetSubVideoSource(ctx context.Context) (nvrxsd.VideoSourceConfiguration, error) {
	return m.GetVideoSource(ctx, m.Camera.SubStreamProfileToken)
}

func (m *ONVIFManager) GetVideoSource(ctx context.Context, token string) (nvrxsd.VideoSourceConfiguration, error) {

	req := media.GetVideoSourceConfiguration {
		ConfigurationToken: xsd.ReferenceToken(token),
	}

	data, err := CallONVIFRequest[nvrmedia.GetVideoSourceConfigurationResponse](m.Device, req)
	if err != nil {
		return nvrxsd.VideoSourceConfiguration{}, fmt.Errorf("failed to fetch video source: %w", err)
	}

	return data.Configuration, nil
}

// func (m *ONVIFManager) GetVideoSource(ctx context.Context, token string) (onvifxml.VideoSourceConfiguration, error) {

// 	req := media.GetVideoSourceConfiguration {
// 		ConfigurationToken: xsd.ReferenceToken(token),
// 	}

// 	res, err := m.Device.CallMethod(req)

// 	var data onvifxml.VideoSourceConfiguration

// 	// res, err := sdk.Call_GetVideoSourceConfiguration(ctx, m.Device, req)
// 	if err != nil {
// 		return data, fmt.Errorf("failed to fetch video source: %w", err)
// 	}

// 	resBytes, err := io.ReadAll(res.Body)

// 	var env onvifxml.Envelope[GetVideoSourceConfigurationResponse]

// 	if err := xml.NewDecoder(bytes.NewReader(resBytes)).Decode(&env); err != nil {
// 		return data, fmt.Errorf("failed to parse response: %w", err)
// 	}

// 	data = env.Body.Content.Configuration

// 	return data, nil
// }