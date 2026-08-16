package ptz

import (
	"github.com/use-go/onvif/xsd"
	// "github.com/use-go/onvif/xsd/onvif"
	onvif "nvr_core/onvif/xml"
)

type Capabilities struct {
	EFlip                       xsd.Boolean `xml:"EFlip,attr"`
	Reverse                     xsd.Boolean `xml:"Reverse,attr"`
	GetCompatibleConfigurations xsd.Boolean `xml:"GetCompatibleConfigurations,attr"`
	MoveStatus                  xsd.Boolean `xml:"MoveStatus,attr"`
	StatusPosition              xsd.Boolean `xml:"StatusPosition,attr"`
}

//PTZ main types

type GetServiceCapabilities struct {
	XMLName string `xml:"GetServiceCapabilities"`
}

type GetServiceCapabilitiesResponse struct {
	Capabilities Capabilities
}

type GetNodes struct {
	XMLName string `xml:"GetNodes"`
}

type GetNodesResponse struct {
	PTZNode onvif.PTZNode
}

type GetNode struct {
	XMLName   string               `xml:"GetNode"`
	NodeToken onvif.ReferenceToken `xml:"NodeToken"`
}

type GetNodeResponse struct {
	PTZNode onvif.PTZNode
}

type GetConfiguration struct {
	XMLName      string               `xml:"GetConfiguration"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetConfigurationResponse struct {
	PTZConfiguration onvif.PTZConfiguration
}

type GetConfigurations struct {
	XMLName string `xml:"GetConfigurations"`
}

type GetConfigurationsResponse struct {
	PTZConfiguration onvif.PTZConfiguration
}

type SetConfiguration struct {
	XMLName          string                 `xml:"SetConfiguration"`
	PTZConfiguration onvif.PTZConfiguration `xml:"PTZConfiguration"`
	ForcePersistence xsd.Boolean            `xml:"ForcePersistence"`
}

type SetConfigurationResponse struct {
}

type GetConfigurationOptions struct {
	XMLName      string               `xml:"GetConfigurationOptions"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetConfigurationOptionsResponse struct {
	PTZConfigurationOptions onvif.PTZConfigurationOptions
}

type SendAuxiliaryCommand struct {
	XMLName       string               `xml:"SendAuxiliaryCommand"`
	ProfileToken  onvif.ReferenceToken `xml:"ProfileToken"`
	AuxiliaryData onvif.AuxiliaryData  `xml:"AuxiliaryData"`
}

type SendAuxiliaryCommandResponse struct {
	AuxiliaryResponse onvif.AuxiliaryData
}

type GetPresets struct {
	XMLName      string               `xml:"GetPresets"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetPresetsResponse struct {
	Preset onvif.PTZPreset
}

type SetPreset struct {
	XMLName      string               `xml:"SetPreset"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	PresetName   xsd.String           `xml:"PresetName"`
	PresetToken  onvif.ReferenceToken `xml:"PresetToken,omitempty"`
}

type SetPresetResponse struct {
	PresetToken onvif.ReferenceToken
}

type RemovePreset struct {
	XMLName      string               `xml:"RemovePreset"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	PresetToken  onvif.ReferenceToken `xml:"PresetToken"`
}

type RemovePresetResponse struct {
}

type GotoPreset struct {
	XMLName      string               `xml:"GotoPreset"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	PresetToken  onvif.ReferenceToken `xml:"PresetToken"`
	Speed        onvif.PTZSpeed       `xml:"Speed"`
}

type GotoPresetResponse struct {
}

type GotoHomePosition struct {
	XMLName      string               `xml:"GotoHomePosition"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	Speed        onvif.PTZSpeed       `xml:"Speed"`
}

type GotoHomePositionResponse struct {
}

type SetHomePosition struct {
	XMLName      string               `xml:"SetHomePosition"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type SetHomePositionResponse struct {
}

type ContinuousMove struct {
	XMLName      string               `xml:"ContinuousMove"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	Velocity     onvif.PTZSpeed       `xml:"Velocity"`
	Timeout      xsd.Duration         `xml:"Timeout"`
}

type ContinuousMoveResponse struct {
}

type RelativeMove struct {
	XMLName      string               `xml:"RelativeMove"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	Translation  onvif.PTZVector      `xml:"Translation"`
	Speed        onvif.PTZSpeed       `xml:"Speed"`
}

type RelativeMoveResponse struct {
}

type GetStatus struct {
	XMLName      string               `xml:"GetStatus"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetStatusResponse struct {
	PTZStatus onvif.PTZStatus
}

type AbsoluteMove struct {
	XMLName      string               `xml:"AbsoluteMove"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	Position     onvif.PTZVector      `xml:"Position"`
	Speed        onvif.PTZSpeed       `xml:"Speed"`
}

type AbsoluteMoveResponse struct {
}

type GeoMove struct {
	XMLName      string               `xml:"GeoMove"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	Target       onvif.GeoLocation    `xml:"Target"`
	Speed        onvif.PTZSpeed       `xml:"Speed"`
	AreaHeight   xsd.Float            `xml:"AreaHeight"`
	AreaWidth    xsd.Float            `xml:"AreaWidth"`
}

type GeoMoveResponse struct {
}

type Stop struct {
	XMLName      string               `xml:"Stop"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	PanTilt      xsd.Boolean          `xml:"PanTilt"`
	Zoom         xsd.Boolean          `xml:"Zoom"`
}

type StopResponse struct {
}

type GetPresetTours struct {
	XMLName      string               `xml:"GetPresetTours"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetPresetToursResponse struct {
	PresetTour onvif.PresetTour
}

type GetPresetTour struct {
	XMLName         string               `xml:"GetPresetTour"`
	ProfileToken    onvif.ReferenceToken `xml:"ProfileToken"`
	PresetTourToken onvif.ReferenceToken `xml:"PresetTourToken"`
}

type GetPresetTourResponse struct {
	PresetTour onvif.PresetTour
}

type GetPresetTourOptions struct {
	XMLName         string               `xml:"GetPresetTourOptions"`
	ProfileToken    onvif.ReferenceToken `xml:"ProfileToken"`
	PresetTourToken onvif.ReferenceToken `xml:"PresetTourToken"`
}

type GetPresetTourOptionsResponse struct {
	Options onvif.PTZPresetTourOptions
}

type CreatePresetTour struct {
	XMLName      string               `xml:"CreatePresetTour"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type CreatePresetTourResponse struct {
	PresetTourToken onvif.ReferenceToken
}

type ModifyPresetTour struct {
	XMLName      string               `xml:"ModifyPresetTour"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
	PresetTour   onvif.PresetTour     `xml:"PresetTour"`
}

type ModifyPresetTourResponse struct {
}

type OperatePresetTour struct {
	XMLName         string                       `xml:"OperatePresetTour"`
	ProfileToken    onvif.ReferenceToken         `xml:"ProfileToken"`
	PresetTourToken onvif.ReferenceToken         `xml:"onvif:PresetTourToken"`
	Operation       onvif.PTZPresetTourOperation `xml:"onvif:Operation"`
}

type OperatePresetTourResponse struct {
}

type RemovePresetTour struct {
	XMLName         string               `xml:"RemovePresetTour"`
	ProfileToken    onvif.ReferenceToken `xml:"ProfileToken"`
	PresetTourToken onvif.ReferenceToken `xml:"PresetTourToken"`
}

type RemovePresetTourResponse struct {
}

type GetCompatibleConfigurations struct {
	XMLName      string               `xml:"GetCompatibleConfigurations"`
	ProfileToken onvif.ReferenceToken `xml:"ProfileToken"`
}

type GetCompatibleConfigurationsResponse struct {
	PTZConfiguration onvif.PTZConfiguration
}
