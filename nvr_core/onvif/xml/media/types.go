package xml

import (
	"github.com/use-go/onvif/xsd"
	onvif "nvr_core/onvif/xml"
)


type Capabilities struct {
	SnapshotUri           bool `xml:"SnapshotUri,attr"`
	Rotation              bool `xml:"Rotation,attr"`
	VideoSourceMode       bool `xml:"VideoSourceMode,attr"`
	OSD                   bool `xml:"OSD,attr"`
	TemporaryOSDText      bool `xml:"TemporaryOSDText,attr"`
	EXICompression        bool `xml:"EXICompression,attr"`
	ProfileCapabilities   ProfileCapabilities
	StreamingCapabilities StreamingCapabilities
}

type ProfileCapabilities struct {
	MaximumNumberOfProfiles int `xml:"MaximumNumberOfProfiles,attr"`
}

type StreamingCapabilities struct {
	RTPMulticast        bool `xml:"RTPMulticast,attr"`
	RTP_TCP             bool `xml:"RTP_TCP,attr"`
	RTP_RTSP_TCP        bool `xml:"RTP_RTSP_TCP,attr"`
	NonAggregateControl bool `xml:"NonAggregateControl,attr"`
	NoRTSPStreaming     bool `xml:"NoRTSPStreaming,attr"`
}

//Media main types

type GetServiceCapabilities struct {
	XMLName string `xml:"GetServiceCapabilities"`
}

type GetServiceCapabilitiesResponse struct {
	XMLName string `xml:"GetServiceCapabilitiesResponse"`
	Capabilities Capabilities
}

type GetVideoSources struct {
	XMLName string `xml:"GetVideoSources"`
}

type GetVideoSourcesResponse struct {
	XMLName string `xml:"GetVideoSourcesResponse"`
	VideoSources onvif.VideoSource
}

type GetAudioSources struct {
	XMLName string `xml:"GetAudioSources"`
}

type GetAudioSourcesResponse struct {
	XMLName string `xml:"GetAudioSourcesResponse"`
	AudioSources onvif.AudioSource
}

type GetAudioOutputs struct {
	XMLName string `xml:"GetAudioOutputs"`
}

type GetAudioOutputsResponse struct {
	XMLName string `xml:"GetAudioOutputsResponse"`
	AudioOutputs onvif.AudioOutput
}

type CreateProfile struct {
	XMLName string               `xml:"CreateProfile"`
	Name    onvif.Name           `xml:"Name"`
	Token   onvif.ReferenceToken `xml:"Token"`
}

type CreateProfileResponse struct {
	Profile onvif.Profile
}

type GetProfile struct {
	XMLName      string               `xml:"GetProfile"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetProfileResponse struct {
	XMLName string `xml:"GetProfileResponse"`
	Profile onvif.Profile
}

type GetProfiles struct {
	XMLName string `xml:"GetProfiles"`
}

type GetProfilesResponse struct {
	XMLName string `xml:"GetProfilesResponse"`
	Profiles []onvif.Profile
}

type AddVideoEncoderConfiguration struct {
	XMLName string `xml:"AddVideoEncoderConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddVideoEncoderConfigurationResponse struct {
}

type RemoveVideoEncoderConfiguration struct {
	XMLName      string               `xml:"RemoveVideoEncoderConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveVideoEncoderConfigurationResponse struct {
}

type AddVideoSourceConfiguration struct {
	XMLName string `xml:"AddVideoSourceConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddVideoSourceConfigurationResponse struct {
}

type RemoveVideoSourceConfiguration struct {
	XMLName      string               `xml:"RemoveVideoSourceConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveVideoSourceConfigurationResponse struct {
}

type AddAudioEncoderConfiguration struct {
	XMLName string `xml:"AddAudioEncoderConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddAudioEncoderConfigurationResponse struct {
}

type RemoveAudioEncoderConfiguration struct {
	XMLName      string               `xml:"RemoveAudioEncoderConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveAudioEncoderConfigurationResponse struct {
}

type AddAudioSourceConfiguration struct {
	XMLName string `xml:"AddAudioSourceConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddAudioSourceConfigurationResponse struct {
}

type RemoveAudioSourceConfiguration struct {
	XMLName      string               `xml:"RemoveAudioSourceConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveAudioSourceConfigurationResponse struct {
}

type AddPTZConfiguration struct {
	XMLName string `xml:"AddPTZConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddPTZConfigurationResponse struct {
}

type RemovePTZConfiguration struct {
	XMLName      string               `xml:"RemovePTZConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemovePTZConfigurationResponse struct {
}

type AddVideoAnalyticsConfiguration struct {
	XMLName string `xml:"AddVideoAnalyticsConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddVideoAnalyticsConfigurationResponse struct {
}

type RemoveVideoAnalyticsConfiguration struct {
	XMLName      string               `xml:"RemoveVideoAnalyticsConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveVideoAnalyticsConfigurationResponse struct {
}

type AddMetadataConfiguration struct {
	XMLName string `xml:"AddMetadataConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddMetadataConfigurationResponse struct {
}

type RemoveMetadataConfiguration struct {
	XMLName      string               `xml:"RemoveMetadataConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveMetadataConfigurationResponse struct {
}

type AddAudioOutputConfiguration struct {
	XMLName string `xml:"AddAudioOutputConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddAudioOutputConfigurationResponse struct {
}

type RemoveAudioOutputConfiguration struct {
	XMLName      string               `xml:"RemoveAudioOutputConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveAudioOutputConfigurationResponse struct {
}

type AddAudioDecoderConfiguration struct {
	XMLName string `xml:"AddAudioDecoderConfiguration"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type AddAudioDecoderConfigurationResponse struct {
}

type RemoveAudioDecoderConfiguration struct {
	XMLName      string               `xml:"RemoveAudioDecoderConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type RemoveAudioDecoderConfigurationResponse struct {
}

type DeleteProfile struct {
	XMLName      string               `xml:"DeleteProfile"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type DeleteProfileResponse struct {
}

type GetVideoSourceConfigurations struct {
	XMLName string `xml:"GetVideoSourceConfigurations"`
}

type GetVideoSourceConfigurationsResponse struct {
	XMLName string `xml:"GetVideoSourceConfigurationsResponse"`
	Configurations onvif.VideoSourceConfiguration
}

type GetVideoEncoderConfigurations struct {
	XMLName string `xml:"GetVideoEncoderConfigurations"`
}

type GetVideoEncoderConfigurationsResponse struct {
	XMLName string `xml:"GetVideoEncoderConfigurationsResponse"`
	Configurations onvif.VideoEncoderConfiguration
}

type GetAudioSourceConfigurations struct {
	XMLName string `xml:"GetAudioSourceConfigurations"`
}

type GetAudioSourceConfigurationsResponse struct {
	XMLName string `xml:"GetAudioSourceConfigurationsResponse"`
	Configurations onvif.AudioSourceConfiguration
}

type GetAudioEncoderConfigurations struct {
	XMLName string `xml:"GetAudioEncoderConfigurations"`
}

type GetAudioEncoderConfigurationsResponse struct {
	XMLName string `xml:"GetAudioEncoderConfigurationsResponse"`
	Configurations onvif.AudioEncoderConfiguration
}

type GetVideoAnalyticsConfigurations struct {
	XMLName string `xml:"GetVideoAnalyticsConfigurations"`
}

type GetVideoAnalyticsConfigurationsResponse struct {
	XMLName string `xml:"GetVideoAnalyticsConfigurationsResponse"`
	Configurations onvif.VideoAnalyticsConfiguration
}

type GetMetadataConfigurations struct {
	XMLName string `xml:"GetMetadataConfigurations"`
}

type GetMetadataConfigurationsResponse struct {
	XMLName string `xml:"GetMetadataConfigurationsResponse"`
	Configurations onvif.MetadataConfiguration
}

type GetAudioOutputConfigurations struct {
	XMLName string `xml:"GetAudioOutputConfigurations"`
}

type GetAudioOutputConfigurationsResponse struct {
	XMLName string `xml:"GetAudioOutputConfigurationsResponse"`
	Configurations onvif.AudioOutputConfiguration
}

type GetAudioDecoderConfigurations struct {
	XMLName string `xml:"GetAudioDecoderConfigurations"`
}

type GetAudioDecoderConfigurationsResponse struct {
	XMLName string `xml:"GetAudioDecoderConfigurationsResponse"`
	Configurations onvif.AudioDecoderConfiguration
}

type GetVideoSourceConfiguration struct {
	XMLName string `xml:"GetVideoSourceConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetVideoSourceConfigurationResponse struct {
	XMLName string `xml:"GetVideoSourceConfigurationResponse"`
	Configuration onvif.VideoSourceConfiguration
}

type GetVideoEncoderConfiguration struct {
	XMLName string `xml:"GetVideoEncoderConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetVideoEncoderConfigurationResponse struct {
	XMLName string `xml:"GetVideoEncoderConfigurationResponse"`
	Configuration onvif.VideoEncoderConfiguration
}

type GetAudioSourceConfiguration struct {
	XMLName string `xml:"GetAudioSourceConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioSourceConfigurationResponse struct {
	XMLName string `xml:"GetAudioSourceConfigurationResponse"`
	Configuration onvif.AudioSourceConfiguration
}

type GetAudioEncoderConfiguration struct {
	XMLName string `xml:"GetAudioEncoderConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioEncoderConfigurationResponse struct {
	XMLName string `xml:"GetAudioEncoderConfigurationResponse"`
	Configuration onvif.AudioEncoderConfiguration
}

type GetVideoAnalyticsConfiguration struct {
	XMLName string `xml:"GetVideoAnalyticsConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetVideoAnalyticsConfigurationResponse struct {
	XMLName string `xml:"GetVideoAnalyticsConfigurationResponse"`
	Configuration onvif.VideoAnalyticsConfiguration
}

type GetMetadataConfiguration struct {
	XMLName string `xml:"GetMetadataConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetMetadataConfigurationResponse struct {
	XMLName string `xml:"GetMetadataConfigurationResponse"`
	Configuration onvif.MetadataConfiguration
}

type GetAudioOutputConfiguration struct {
	XMLName string `xml:"GetAudioOutputConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioOutputConfigurationResponse struct {
	XMLName string `xml:"GetAudioOutputConfigurationResponse"`
	Configuration onvif.AudioOutputConfiguration
}

type GetAudioDecoderConfiguration struct {
	XMLName string `xml:"GetAudioDecoderConfiguration"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioDecoderConfigurationResponse struct {
	XMLName string `xml:"GetAudioDecoderConfigurationResponse"`
	Configuration onvif.AudioDecoderConfiguration
}

type GetCompatibleVideoEncoderConfigurations struct {
	XMLName      string               `xml:"GetCompatibleVideoEncoderConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleVideoEncoderConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleVideoEncoderConfigurationsResponse"`
	Configurations onvif.VideoEncoderConfiguration
}

type GetCompatibleVideoSourceConfigurations struct {
	XMLName      string               `xml:"GetCompatibleVideoSourceConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleVideoSourceConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleVideoSourceConfigurationsResponse"`
	Configurations onvif.VideoSourceConfiguration
}

type GetCompatibleAudioEncoderConfigurations struct {
	XMLName      string               `xml:"GetCompatibleAudioEncoderConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleAudioEncoderConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleAudioEncoderConfigurationsResponse"`
	Configurations onvif.AudioEncoderConfiguration
}

type GetCompatibleAudioSourceConfigurations struct {
	XMLName      string               `xml:"GetCompatibleAudioSourceConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleAudioSourceConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleAudioSourceConfigurationsResponse"`
	Configurations onvif.AudioSourceConfiguration
}

type GetCompatibleVideoAnalyticsConfigurations struct {
	XMLName      string               `xml:"GetCompatibleVideoAnalyticsConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleVideoAnalyticsConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleVideoAnalyticsConfigurationsResponse"`
	Configurations onvif.VideoAnalyticsConfiguration
}

type GetCompatibleMetadataConfigurations struct {
	XMLName      string               `xml:"GetCompatibleMetadataConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleMetadataConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleMetadataConfigurationsResponse"`
	Configurations onvif.MetadataConfiguration
}

type GetCompatibleAudioOutputConfigurations struct {
	XMLName      string               `xml:"GetCompatibleAudioOutputConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleAudioOutputConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleAudioOutputConfigurationsResponse"`
	Configurations onvif.AudioOutputConfiguration
}

type GetCompatibleAudioDecoderConfigurations struct {
	XMLName      string               `xml:"GetCompatibleAudioDecoderConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleAudioDecoderConfigurationsResponse struct {
	XMLName string `xml:"GetCompatibleAudioDecoderConfigurationsResponse"`
	Configurations onvif.AudioDecoderConfiguration
}

type SetVideoSourceConfiguration struct {
	XMLName          string                         `xml:"SetVideoSourceConfiguration"`
	Configuration    onvif.VideoSourceConfiguration `xml:"Configuration"`
	ForcePersistence xsd.Boolean                    `xml:"ForcePersistence"`
}

type SetVideoSourceConfigurationResponse struct {
}

type SetVideoEncoderConfiguration struct {
	XMLName          string                          `xml:"SetVideoEncoderConfiguration"`
	Configuration    onvif.VideoEncoderConfiguration `xml:"Configuration"`
	ForcePersistence xsd.Boolean                     `xml:"ForcePersistence"`
}

type SetVideoEncoderConfigurationResponse struct {
}

type SetAudioSourceConfiguration struct {
	XMLName          string                         `xml:"SetAudioSourceConfiguration"`
	Configuration    onvif.AudioSourceConfiguration `xml:"Configuration"`
	ForcePersistence xsd.Boolean                    `xml:"ForcePersistence"`
}

type SetAudioSourceConfigurationResponse struct {
}

type SetAudioEncoderConfiguration struct {
	XMLName          string                          `xml:"SetAudioEncoderConfiguration"`
	Configuration    onvif.AudioEncoderConfiguration `xml:"Configuration"`
	ForcePersistence xsd.Boolean                     `xml:"ForcePersistence"`
}

type SetAudioEncoderConfigurationResponse struct {
}

type SetVideoAnalyticsConfiguration struct {
	XMLName          string                            `xml:"SetVideoAnalyticsConfiguration"`
	Configuration    onvif.VideoAnalyticsConfiguration `xml:"Configuration"`
	ForcePersistence bool                              `xml:"ForcePersistence"`
}

type SetVideoAnalyticsConfigurationResponse struct {
}

type SetMetadataConfiguration struct {
	XMLName          string                      `xml:"GetDeviceInformation"`
	Configuration    onvif.MetadataConfiguration `xml:"Configuration"`
	ForcePersistence xsd.Boolean                 `xml:"ForcePersistence"`
}

type SetMetadataConfigurationResponse struct {
}

type SetAudioOutputConfiguration struct {
	XMLName          string                         `xml:"SetAudioOutputConfiguration"`
	Configuration    onvif.AudioOutputConfiguration `xml:"Configuration"`
	ForcePersistence bool                           `xml:"ForcePersistence"`
}

type SetAudioOutputConfigurationResponse struct {
}

type SetAudioDecoderConfiguration struct {
	XMLName          string                          `xml:"SetAudioDecoderConfiguration"`
	Configuration    onvif.AudioDecoderConfiguration `xml:"Configuration"`
	ForcePersistence xsd.Boolean                     `xml:"ForcePersistence"`
}

type SetAudioDecoderConfigurationResponse struct {
}

type GetVideoSourceConfigurationOptions struct {
	XMLName string `xml:"GetVideoSourceConfigurationOptions"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetVideoSourceConfigurationOptionsResponse struct {
	XMLName string `xml:"GetVideoSourceConfigurationOptionsResponse"`
	Options onvif.VideoSourceConfigurationOptions
}

type GetVideoEncoderConfigurationOptions struct {
	XMLName string `xml:"GetVideoEncoderConfigurationOptions"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetVideoEncoderConfigurationOptionsResponse struct {
	XMLName string `xml:"GetVideoEncoderConfigurationOptionsResponse"`
	Options onvif.VideoEncoderConfigurationOptions
}

type GetAudioSourceConfigurationOptions struct {
	XMLName string `xml:"GetAudioSourceConfigurationOptions"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioSourceConfigurationOptionsResponse struct {
	XMLName string `xml:"GetAudioSourceConfigurationOptionsResponse"`
	Options onvif.AudioSourceConfigurationOptions
}

type GetAudioEncoderConfigurationOptions struct {
	XMLName string `xml:"GetAudioEncoderConfigurationOptions"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioEncoderConfigurationOptionsResponse struct {
	XMLName string `xml:"GetAudioEncoderConfigurationOptionsResponse"`
	Options onvif.AudioEncoderConfigurationOptions
}

type GetMetadataConfigurationOptions struct {
	XMLName string `xml:"GetMetadataConfigurationOptions"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetMetadataConfigurationOptionsResponse struct {
	XMLName string `xml:"GetMetadataConfigurationOptionsResponse"`
	Options onvif.MetadataConfigurationOptions
}

type GetAudioOutputConfigurationOptions struct {
	XMLName string `xml:"GetAudioOutputConfigurationOptions"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioOutputConfigurationOptionsResponse struct {
	XMLName string `xml:"GetAudioOutputConfigurationOptionsResponse"`
	Options onvif.AudioOutputConfigurationOptions
}

type GetAudioDecoderConfigurationOptions struct {
	XMLName string `xml:"GetAudioDecoderConfigurationOptions"`
	ProfileToken       onvif.ReferenceToken `xml:"ProfileToken"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetAudioDecoderConfigurationOptionsResponse struct {
	XMLName string `xml:"GetAudioDecoderConfigurationOptionsResponse"`
	Options onvif.AudioDecoderConfigurationOptions
}

type GetGuaranteedNumberOfVideoEncoderInstances struct {
	XMLName string `xml:"GetGuaranteedNumberOfVideoEncoderInstances"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetGuaranteedNumberOfVideoEncoderInstancesResponse struct {
	XMLName string `xml:"GetGuaranteedNumberOfVideoEncoderInstancesResponse"`
	TotalNumber int
	JPEG        int
	H264        int
	MPEG4       int
}

type GetStreamUri struct {
	XMLName      string               `xml:"GetStreamUri"`
	StreamSetup  onvif.StreamSetup    `xml:"StreamSetup"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetStreamUriResponse struct {
	XMLName string `xml:"GetStreamUriResponse"`
	MediaUri onvif.MediaUri
}

type StartMulticastStreaming struct {
	XMLName      string               `xml:"StartMulticastStreaming"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type StartMulticastStreamingResponse struct {
}

type StopMulticastStreaming struct {
	XMLName      string               `xml:"StopMulticastStreaming"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type StopMulticastStreamingResponse struct {
}

type SetSynchronizationPoint struct {
	XMLName      string               `xml:"SetSynchronizationPoint"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type SetSynchronizationPointResponse struct {
}

type GetSnapshotUri struct {
	XMLName      string               `xml:"GetSnapshotUri"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetSnapshotUriResponse struct {
	XMLName string `xml:"GetSnapshotUriResponse"`
	MediaUri onvif.MediaUri
}

type GetVideoSourceModes struct {
	XMLName          string               `xml:"GetVideoSourceModes"`
	VideoSourceToken onvif.ReferenceToken `xml:"VideoSourceToken"`
}

type GetVideoSourceModesResponse struct {
	XMLName string `xml:"GetVideoSourceModesResponse"`
	VideoSourceModes onvif.VideoSourceMode
}

type SetVideoSourceMode struct {
	XMLName              string               `xml:"SetVideoSourceMode"`
	VideoSourceToken     onvif.ReferenceToken `xml:"VideoSourceToken"`
	VideoSourceModeToken onvif.ReferenceToken `xml:"VideoSourceModeToken"`
}

type SetVideoSourceModeResponse struct {
	Reboot bool
}

type GetOSDs struct {
	XMLName string `xml:"GetOSDs"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetOSDsResponse struct {
	XMLName string `xml:"struct"`
	OSDs onvif.OSDConfiguration
}

type GetOSD struct {
	XMLName  string               `xml:"GetOSD"`
	OSDToken onvif.ReferenceToken `xml:"OSDToken"`
}

type GetOSDResponse struct {
	XMLName string `xml:"struct"`
	OSD onvif.OSDConfiguration
}

type GetOSDOptions struct {
	XMLName string `xml:"GetOSDOptions"`
	ConfigurationToken onvif.ReferenceToken `xml:"ConfigurationToken"`
}

type GetOSDOptionsResponse struct {
	XMLName string `xml:"GetOSDOptionsResponse"`
	OSDOptions onvif.OSDConfigurationOptions
}

type SetOSD struct {
	XMLName string                 `xml:"SetOSD"`
	OSD     onvif.OSDConfiguration `xml:"OSD"`
}

type SetOSDResponse struct {
}

type CreateOSD struct {
	XMLName string                 `xml:"CreateOSD"`
	OSD     onvif.OSDConfiguration `xml:"OSD"`
}

type CreateOSDResponse struct {
	OSDToken onvif.ReferenceToken
}

type DeleteOSD struct {
	XMLName  string               `xml:"DeleteOSD"`
	OSDToken onvif.ReferenceToken `xml:"OSDToken"`
}

type DeleteOSDResponse struct {
}
