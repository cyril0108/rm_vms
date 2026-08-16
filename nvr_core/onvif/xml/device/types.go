package device

import (
	"github.com/use-go/onvif/xsd"
	onvif "nvr_core/onvif/xml"
)

type Service struct {
	Namespace xsd.AnyURI
	XAddr     xsd.AnyURI
	Capabilities
	Version onvif.OnvifVersion
}

type Capabilities struct {
	Any string
}

type DeviceServiceCapabilities struct {
	Network  NetworkCapabilities
	Security SecurityCapabilities
	System   SystemCapabilities
	Misc     MiscCapabilities
}

type NetworkCapabilities struct {
	IPFilter            xsd.Boolean `xml:"IPFilter,attr"`
	ZeroConfiguration   xsd.Boolean `xml:"ZeroConfiguration,attr"`
	IPVersion6          xsd.Boolean `xml:"IPVersion6,attr"`
	DynDNS              xsd.Boolean `xml:"DynDNS,attr"`
	Dot11Configuration  xsd.Boolean `xml:"Dot11Configuration,attr"`
	Dot1XConfigurations int         `xml:"Dot1XConfigurations,attr"`
	HostnameFromDHCP    xsd.Boolean `xml:"HostnameFromDHCP,attr"`
	NTP                 int         `xml:"NTP,attr"`
	DHCPv6              xsd.Boolean `xml:"DHCPv6,attr"`
}

type SecurityCapabilities struct {
	TLS1_0               xsd.Boolean    `xml:"TLS1_0,attr"`
	TLS1_1               xsd.Boolean    `xml:"TLS1_1,attr"`
	TLS1_2               xsd.Boolean    `xml:"TLS1_2,attr"`
	OnboardKeyGeneration xsd.Boolean    `xml:"OnboardKeyGeneration,attr"`
	AccessPolicyConfig   xsd.Boolean    `xml:"AccessPolicyConfig,attr"`
	DefaultAccessPolicy  xsd.Boolean    `xml:"DefaultAccessPolicy,attr"`
	Dot1X                xsd.Boolean    `xml:"Dot1X,attr"`
	RemoteUserHandling   xsd.Boolean    `xml:"RemoteUserHandling,attr"`
	X_509Token           xsd.Boolean    `xml:"X_509Token,attr"`
	SAMLToken            xsd.Boolean    `xml:"SAMLToken,attr"`
	KerberosToken        xsd.Boolean    `xml:"KerberosToken,attr"`
	UsernameToken        xsd.Boolean    `xml:"UsernameToken,attr"`
	HttpDigest           xsd.Boolean    `xml:"HttpDigest,attr"`
	RELToken             xsd.Boolean    `xml:"RELToken,attr"`
	SupportedEAPMethods  EAPMethodTypes `xml:"SupportedEAPMethods,attr"`
	MaxUsers             int            `xml:"MaxUsers,attr"`
	MaxUserNameLength    int            `xml:"MaxUserNameLength,attr"`
	MaxPasswordLength    int            `xml:"MaxPasswordLength,attr"`
}

type EAPMethodTypes struct {
	Types []int
}

type SystemCapabilities struct {
	DiscoveryResolve         xsd.Boolean          `xml:"DiscoveryResolve,attr"`
	DiscoveryBye             xsd.Boolean          `xml:"DiscoveryBye,attr"`
	RemoteDiscovery          xsd.Boolean          `xml:"RemoteDiscovery,attr"`
	SystemBackup             xsd.Boolean          `xml:"SystemBackup,attr"`
	SystemLogging            xsd.Boolean          `xml:"SystemLogging,attr"`
	FirmwareUpgrade          xsd.Boolean          `xml:"FirmwareUpgrade,attr"`
	HttpFirmwareUpgrade      xsd.Boolean          `xml:"HttpFirmwareUpgrade,attr"`
	HttpSystemBackup         xsd.Boolean          `xml:"HttpSystemBackup,attr"`
	HttpSystemLogging        xsd.Boolean          `xml:"HttpSystemLogging,attr"`
	HttpSupportInformation   xsd.Boolean          `xml:"HttpSupportInformation,attr"`
	StorageConfiguration     xsd.Boolean          `xml:"StorageConfiguration,attr"`
	MaxStorageConfigurations int                  `xml:"MaxStorageConfigurations,attr"`
	GeoLocationEntries       int                  `xml:"GeoLocationEntries,attr"`
	AutoGeo                  onvif.StringAttrList `xml:"AutoGeo,attr"`
}

type MiscCapabilities struct {
	AuxiliaryCommands onvif.StringAttrList `xml:"AuxiliaryCommands,attr"`
}

type StorageConfiguration struct {
	onvif.DeviceEntity
	Data StorageConfigurationData `xml:"Data"`
}

type StorageConfigurationData struct {
	Type       xsd.String     `xml:"type,attr"`
	LocalPath  xsd.AnyURI     `xml:"LocalPath"`
	StorageUri xsd.AnyURI     `xml:"StorageUri"`
	User       UserCredential `xml:"User"`
	Extension  xsd.AnyURI     `xml:"Extension"`
}

type UserCredential struct {
	UserName  xsd.String  `xml:"UserName"`
	Password  xsd.String  `xml:"Password"`
	Extension xsd.AnyType `xml:"Extension"`
}

//Device main types

type GetServices struct {
	XMLName           string      `xml:"GetServices"`
	IncludeCapability xsd.Boolean `xml:"IncludeCapability"`
}

type GetServicesResponse struct {
	XMLName string `xml:"GetServicesResponse"`
	Service Service
}

type GetServiceCapabilities struct {
	XMLName string `xml:"GetServiceCapabilities"`
}

type GetServiceCapabilitiesResponse struct {
	XMLName string `xml:"GetServiceCapabilitiesResponse"`
	Capabilities DeviceServiceCapabilities
}

type GetDeviceInformation struct {
	XMLName string `xml:"GetDeviceInformation"`
}

type GetDeviceInformationResponse struct {
	XMLName string `xml:"GetDeviceInformationResponse"`
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareId      string
}

type SetSystemDateAndTime struct {
	XMLName         string                `xml:"SetSystemDateAndTime"`
	DateTimeType    onvif.SetDateTimeType `xml:"DateTimeType"`
	DaylightSavings xsd.Boolean           `xml:"DaylightSavings"`
	TimeZone        onvif.TimeZone        `xml:"TimeZone"`
	UTCDateTime     onvif.DateTime        `xml:"UTCDateTime"`
}

type SetSystemDateAndTimeResponse struct {
}

type GetSystemDateAndTime struct {
	XMLName string `xml:"GetSystemDateAndTime"`
}

type GetSystemDateAndTimeResponse struct {
	XMLName string `xml:"GetSystemDateAndTimeResponse"`
	SystemDateAndTime onvif.SystemDateTime
}

type SetSystemFactoryDefault struct {
	XMLName        string                   `xml:"SetSystemFactoryDefault"`
	FactoryDefault onvif.FactoryDefaultType `xml:"FactoryDefault"`
}

type SetSystemFactoryDefaultResponse struct {
}

type UpgradeSystemFirmware struct {
	XMLName  string               `xml:"UpgradeSystemFirmware"`
	Firmware onvif.AttachmentData `xml:"Firmware"`
}

type UpgradeSystemFirmwareResponse struct {
	Message string
}

type SystemReboot struct {
	XMLName string `xml:"SystemReboot"`
}

type SystemRebootResponse struct {
	Message string
}

//TODO: one or more repetitions
type RestoreSystem struct {
	XMLName     string           `xml:"RestoreSystem"`
	BackupFiles onvif.BackupFile `xml:"BackupFiles"`
}

type RestoreSystemResponse struct {
}

type GetSystemBackup struct {
	XMLName string `xml:"GetSystemBackup"`
}

type GetSystemBackupResponse struct {
	XMLName string `xml:"GetSystemBackupResponse"`
	BackupFiles onvif.BackupFile
}

type GetSystemLog struct {
	XMLName string              `xml:"GetSystemLog"`
	LogType onvif.SystemLogType `xml:"LogType"`
}

type GetSystemLogResponse struct {
	XMLName string `xml:"GetSystemLogResponse"`
	SystemLog onvif.SystemLog
}

type GetSystemSupportInformation struct {
	XMLName string `xml:"GetSystemSupportInformation"`
}

type GetSystemSupportInformationResponse struct {
	XMLName string `xml:"GetSystemSupportInformationResponse"`
	SupportInformation onvif.SupportInformation
}

type GetScopes struct {
	XMLName string `xml:"GetScopes"`
}

type GetScopesResponse struct {
	XMLName string `xml:"GetScopesResponse"`
	Scopes onvif.Scope
}

//TODO: one or more scopes
type SetScopes struct {
	XMLName string     `xml:"SetScopes"`
	Scopes  xsd.AnyURI `xml:"Scopes"`
}

type SetScopesResponse struct {
}

//TODO: list of scopes
type AddScopes struct {
	XMLName   string     `xml:"AddScopes"`
	ScopeItem xsd.AnyURI `xml:"ScopeItem"`
}

type AddScopesResponse struct {
}

//TODO: One or more repetitions
type RemoveScopes struct {
	XMLName   string     `xml:"RemoveScopes"`
	ScopeItem xsd.AnyURI `xml:"onvif:ScopeItem"`
}

type RemoveScopesResponse struct {
	ScopeItem xsd.AnyURI
}

type GetDiscoveryMode struct {
	XMLName string `xml:"GetDiscoveryMode"`
}

type GetDiscoveryModeResponse struct {
	XMLName string `xml:"GetDiscoveryModeResponse"`
	DiscoveryMode onvif.DiscoveryMode
}

type SetDiscoveryMode struct {
	XMLName       string              `xml:"SetDiscoveryMode"`
	DiscoveryMode onvif.DiscoveryMode `xml:"DiscoveryMode"`
}

type SetDiscoveryModeResponse struct {
}

type GetRemoteDiscoveryMode struct {
	XMLName string `xml:"GetRemoteDiscoveryMode"`
}

type GetRemoteDiscoveryModeResponse struct {
	XMLName string `xml:"GetRemoteDiscoveryModeResponse"`
	RemoteDiscoveryMode onvif.DiscoveryMode
}

type SetRemoteDiscoveryMode struct {
	XMLName             string              `xml:"SetRemoteDiscoveryMode"`
	RemoteDiscoveryMode onvif.DiscoveryMode `xml:"RemoteDiscoveryMode"`
}

type SetRemoteDiscoveryModeResponse struct {
}

type GetDPAddresses struct {
	XMLName string `xml:"GetDPAddresses"`
}

type GetDPAddressesResponse struct {
	XMLName string `xml:"GetDPAddressesResponse"`
	DPAddress onvif.NetworkHost
}

type SetDPAddresses struct {
	XMLName   string            `xml:"SetDPAddresses"`
	DPAddress onvif.NetworkHost `xml:"DPAddress"`
}

type SetDPAddressesResponse struct {
}

type GetEndpointReference struct {
	XMLName string `xml:"GetEndpointReference"`
}

type GetEndpointReferenceResponse struct {
	XMLName string `xml:"GetEndpointReferenceResponse"`
	GUID string
}

type GetRemoteUser struct {
	XMLName string `xml:"GetRemoteUser"`
}

type GetRemoteUserResponse struct {
	XMLName string `xml:"GetRemoteUserResponse"`
	RemoteUser onvif.RemoteUser
}

type SetRemoteUser struct {
	XMLName    string           `xml:"SetRemoteUser"`
	RemoteUser onvif.RemoteUser `xml:"RemoteUser"`
}

type SetRemoteUserResponse struct {
}

type GetUsers struct {
	XMLName string `xml:"GetUsers"`
}

type GetUsersResponse struct {
	XMLName string `xml:"GetUsersResponse"`
	User onvif.User
}

//TODO: List of users
type CreateUsers struct {
	XMLName string     `xml:"CreateUsers"`
	User    onvif.User `xml:"User,omitempty"`
}

type CreateUsersResponse struct {
}

//TODO: one or more Username
type DeleteUsers struct {
	XMLName  xsd.String `xml:"DeleteUsers"`
	Username xsd.String `xml:"Username"`
}

type DeleteUsersResponse struct {
}

type SetUser struct {
	XMLName string     `xml:"SetUser"`
	User    onvif.User `xml:"User"`
}

type SetUserResponse struct {
}

type GetWsdlUrl struct {
	XMLName string `xml:"GetWsdlUrl"`
}

type GetWsdlUrlResponse struct {
	XMLName string `xml:"GetWsdlUrlResponse"`
	WsdlUrl xsd.AnyURI
}

type GetCapabilities struct {
	XMLName  string                   `xml:"GetCapabilities"`
	Category onvif.CapabilityCategory `xml:"Category"`
}

type GetCapabilitiesResponse struct {
	XMLName string `xml:"GetCapabilitiesResponse"`
	Capabilities onvif.Capabilities
}

type GetHostname struct {
	XMLName string `xml:"GetHostname"`
}

type GetHostnameResponse struct {
	XMLName string `xml:"GetHostnameResponse"`
	HostnameInformation onvif.HostnameInformation
}

type SetHostname struct {
	XMLName string    `xml:"SetHostname"`
	Name    xsd.Token `xml:"Name"`
}

type SetHostnameResponse struct {
}

type SetHostnameFromDHCP struct {
	XMLName  string      `xml:"SetHostnameFromDHCP"`
	FromDHCP xsd.Boolean `xml:"FromDHCP"`
}

type SetHostnameFromDHCPResponse struct {
	RebootNeeded xsd.Boolean
}

type GetDNS struct {
	XMLName string `xml:"GetDNS"`
}

type GetDNSResponse struct {
	XMLName string `xml:"GetDNSResponse"`
	DNSInformation onvif.DNSInformation
}

type SetDNS struct {
	XMLName      string          `xml:"SetDNS"`
	FromDHCP     xsd.Boolean     `xml:"FromDHCP"`
	SearchDomain xsd.Token       `xml:"SearchDomain"`
	DNSManual    onvif.IPAddress `xml:"DNSManual"`
}

type SetDNSResponse struct {
}

type GetNTP struct {
	XMLName string `xml:"GetNTP"`
}

type GetNTPResponse struct {
	XMLName string `xml:"GetNTPResponse"`
	NTPInformation onvif.NTPInformation
}

type SetNTP struct {
	XMLName   string            `xml:"SetNTP"`
	FromDHCP  xsd.Boolean       `xml:"FromDHCP"`
	NTPManual onvif.NetworkHost `xml:"NTPManual"`
}

type SetNTPResponse struct {
}

type GetDynamicDNS struct {
	XMLName string `xml:"GetDynamicDNS"`
}

type GetDynamicDNSResponse struct {
	XMLName string `xml:"GetDynamicDNSResponse"`
	DynamicDNSInformation onvif.DynamicDNSInformation
}

type SetDynamicDNS struct {
	XMLName string               `xml:"SetDynamicDNS"`
	Type    onvif.DynamicDNSType `xml:"Type"`
	Name    onvif.DNSName        `xml:"Name"`
	TTL     xsd.Duration         `xml:"TTL"`
}

type SetDynamicDNSResponse struct {
}

type GetNetworkInterfaces struct {
	XMLName string `xml:"GetNetworkInterfaces"`
}

type GetNetworkInterfacesResponse struct {
	XMLName string `xml:"GetNetworkInterfacesResponse"`
	NetworkInterfaces onvif.NetworkInterface
}

type SetNetworkInterfaces struct {
	XMLName          string                                 `xml:"SetNetworkInterfaces"`
	InterfaceToken   onvif.ReferenceToken                   `xml:"InterfaceToken"`
	NetworkInterface onvif.NetworkInterfaceSetConfiguration `xml:"NetworkInterface"`
}

type SetNetworkInterfacesResponse struct {
	RebootNeeded xsd.Boolean
}

type GetNetworkProtocols struct {
	XMLName string `xml:"GetNetworkProtocols"`
}

type GetNetworkProtocolsResponse struct {
	XMLName string `xml:"GetNetworkProtocolsResponse"`
	NetworkProtocols onvif.NetworkProtocol
}

type SetNetworkProtocols struct {
	XMLName          string                `xml:"SetNetworkProtocols"`
	NetworkProtocols onvif.NetworkProtocol `xml:"NetworkProtocols"`
}

type SetNetworkProtocolsResponse struct {
}

type GetNetworkDefaultGateway struct {
	XMLName string `xml:"GetNetworkDefaultGateway"`
}

type GetNetworkDefaultGatewayResponse struct {
	XMLName string `xml:"GetNetworkDefaultGatewayResponse"`
	NetworkGateway onvif.NetworkGateway
}

type SetNetworkDefaultGateway struct {
	XMLName     string            `xml:"SetNetworkDefaultGateway"`
	IPv4Address onvif.IPv4Address `xml:"IPv4Address"`
	IPv6Address onvif.IPv6Address `xml:"IPv6Address"`
}

type SetNetworkDefaultGatewayResponse struct {
}

type GetZeroConfiguration struct {
	XMLName string `xml:"GetZeroConfiguration"`
}

type GetZeroConfigurationResponse struct {
	XMLName string `xml:"GetZeroConfigurationResponse"`
	ZeroConfiguration onvif.NetworkZeroConfiguration
}

type SetZeroConfiguration struct {
	XMLName        string               `xml:"SetZeroConfiguration"`
	InterfaceToken onvif.ReferenceToken `xml:"InterfaceToken"`
	Enabled        xsd.Boolean          `xml:"Enabled"`
}

type SetZeroConfigurationResponse struct {
}

type GetIPAddressFilter struct {
	XMLName string `xml:"GetIPAddressFilter"`
}

type GetIPAddressFilterResponse struct {
	XMLName string `xml:"GetIPAddressFilterResponse"`
	IPAddressFilter onvif.IPAddressFilter
}

type SetIPAddressFilter struct {
	XMLName         string                `xml:"SetIPAddressFilter"`
	IPAddressFilter onvif.IPAddressFilter `xml:"IPAddressFilter"`
}

type SetIPAddressFilterResponse struct {
}

//This operation adds an IP filter address to a device.
//If the device supports device access control based on
//IP filtering rules (denied or accepted ranges of IP addresses),
//the device shall support adding of IP filtering addresses through
//the AddIPAddressFilter command.
type AddIPAddressFilter struct {
	XMLName         string                `xml:"AddIPAddressFilter"`
	IPAddressFilter onvif.IPAddressFilter `xml:"IPAddressFilter"`
}

type AddIPAddressFilterResponse struct {
}

type RemoveIPAddressFilter struct {
	XMLName         string                `xml:"RemoveIPAddressFilter"`
	IPAddressFilter onvif.IPAddressFilter `xml:"onvif:IPAddressFilter"`
}

type RemoveIPAddressFilterResponse struct {
}

type GetAccessPolicy struct {
	XMLName string `xml:"GetAccessPolicy"`
}

type GetAccessPolicyResponse struct {
	XMLName string `xml:"GetAccessPolicyResponse"`
	PolicyFile onvif.BinaryData
}

type SetAccessPolicy struct {
	XMLName    string           `xml:"SetAccessPolicy"`
	PolicyFile onvif.BinaryData `xml:"PolicyFile"`
}

type SetAccessPolicyResponse struct {
}

type CreateCertificate struct {
	XMLName        string       `xml:"CreateCertificate"`
	CertificateID  xsd.Token    `xml:"CertificateID,omitempty"`
	Subject        string       `xml:"Subject,omitempty"`
	ValidNotBefore xsd.DateTime `xml:"ValidNotBefore,omitempty"`
	ValidNotAfter  xsd.DateTime `xml:"ValidNotAfter,omitempty"`
}

type CreateCertificateResponse struct {
	NvtCertificate onvif.Certificate
}

type GetCertificates struct {
	XMLName string `xml:"GetCertificates"`
}

type GetCertificatesResponse struct {
	XMLName string `xml:"GetCertificatesResponse"`
	NvtCertificate onvif.Certificate
}

type GetCertificatesStatus struct {
	XMLName string `xml:"GetCertificatesStatus"`
}

type GetCertificatesStatusResponse struct {
	XMLName string `xml:"GetCertificatesStatusResponse"`
	CertificateStatus onvif.CertificateStatus
}

type SetCertificatesStatus struct {
	XMLName           string                  `xml:"SetCertificatesStatus"`
	CertificateStatus onvif.CertificateStatus `xml:"CertificateStatus"`
}

type SetCertificatesStatusResponse struct {
}

//TODO: List of CertificateID
type DeleteCertificates struct {
	XMLName       string    `xml:"DeleteCertificates"`
	CertificateID xsd.Token `xml:"CertificateID"`
}

type DeleteCertificatesResponse struct {
}

//TODO: Откуда onvif:data = cid:21312413412
type GetPkcs10Request struct {
	XMLName       string           `xml:"GetPkcs10Request"`
	CertificateID xsd.Token        `xml:"CertificateID"`
	Subject       xsd.String       `xml:"Subject"`
	Attributes    onvif.BinaryData `xml:"Attributes"`
}

type GetPkcs10RequestResponse struct {
	XMLName string `xml:"GetPkcs10RequestResponse"`
	Pkcs10Request onvif.BinaryData
}

//TODO: one or more NTVCertificate
type LoadCertificates struct {
	XMLName        string            `xml:"LoadCertificates"`
	NVTCertificate onvif.Certificate `xml:"NVTCertificate"`
}

type LoadCertificatesResponse struct {
}

type GetClientCertificateMode struct {
	XMLName string `xml:"GetClientCertificateMode"`
}

type GetClientCertificateModeResponse struct {
	XMLName string `xml:"GetClientCertificateModeResponse"`
	Enabled xsd.Boolean
}

type SetClientCertificateMode struct {
	XMLName string      `xml:"SetClientCertificateMode"`
	Enabled xsd.Boolean `xml:"Enabled"`
}

type SetClientCertificateModeResponse struct {
}

type GetRelayOutputs struct {
	XMLName string `xml:"GetRelayOutputs"`
}

type GetRelayOutputsResponse struct {
	XMLName string `xml:"GetRelayOutputsResponse"`
	RelayOutputs onvif.RelayOutput
}

type SetRelayOutputSettings struct {
	XMLName          string                    `xml:"SetRelayOutputSettings"`
	RelayOutputToken onvif.ReferenceToken      `xml:"RelayOutputToken"`
	Properties       onvif.RelayOutputSettings `xml:"Properties"`
}

type SetRelayOutputSettingsResponse struct {
}

type SetRelayOutputState struct {
	XMLName          string                  `xml:"SetRelayOutputState"`
	RelayOutputToken onvif.ReferenceToken    `xml:"RelayOutputToken"`
	LogicalState     onvif.RelayLogicalState `xml:"LogicalState"`
}

type SetRelayOutputStateResponse struct {
}

type SendAuxiliaryCommand struct {
	XMLName          string              `xml:"SendAuxiliaryCommand"`
	AuxiliaryCommand onvif.AuxiliaryData `xml:"AuxiliaryCommand"`
}

type SendAuxiliaryCommandResponse struct {
	AuxiliaryCommandResponse onvif.AuxiliaryData
}

type GetCACertificates struct {
	XMLName string `xml:"GetCACertificates"`
}

type GetCACertificatesResponse struct {
	XMLName string `xml:"GetCACertificatesResponse"`
	CACertificate onvif.Certificate
}

//TODO: one or more CertificateWithPrivateKey
type LoadCertificateWithPrivateKey struct {
	XMLName                   string                          `xml:"LoadCertificateWithPrivateKey"`
	CertificateWithPrivateKey onvif.CertificateWithPrivateKey `xml:"CertificateWithPrivateKey"`
}

type LoadCertificateWithPrivateKeyResponse struct {
}

type GetCertificateInformation struct {
	XMLName       string    `xml:"GetCertificateInformation"`
	CertificateID xsd.Token `xml:"CertificateID"`
}

type GetCertificateInformationResponse struct {
	XMLName string `xml:"GetCertificateInformationResponse"`
	CertificateInformation onvif.CertificateInformation
}

type LoadCACertificates struct {
	XMLName       string            `xml:"LoadCACertificates"`
	CACertificate onvif.Certificate `xml:"CACertificate"`
}

type LoadCACertificatesResponse struct {
}

type CreateDot1XConfiguration struct {
	XMLName            string                   `xml:"CreateDot1XConfiguration"`
	Dot1XConfiguration onvif.Dot1XConfiguration `xml:"Dot1XConfiguration"`
}

type CreateDot1XConfigurationResponse struct {
}

type SetDot1XConfiguration struct {
	XMLName            string                   `xml:"SetDot1XConfiguration"`
	Dot1XConfiguration onvif.Dot1XConfiguration `xml:"Dot1XConfiguration"`
}

type SetDot1XConfigurationResponse struct {
}

type GetDot1XConfiguration struct {
	XMLName                 string               `xml:"GetDot1XConfiguration"`
	Dot1XConfigurationToken onvif.ReferenceToken `xml:"Dot1XConfigurationToken"`
}

type GetDot1XConfigurationResponse struct {
	XMLName string `xml:"GetDot1XConfigurationResponse"`
	Dot1XConfiguration onvif.Dot1XConfiguration
}

type GetDot1XConfigurations struct {
	XMLName string `xml:"GetDot1XConfigurations"`
}

type GetDot1XConfigurationsResponse struct {
	XMLName string `xml:"GetDot1XConfigurationsResponse"`
	Dot1XConfiguration onvif.Dot1XConfiguration
}

//TODO: Zero or more Dot1XConfigurationToken
type DeleteDot1XConfiguration struct {
	XMLName                 string               `xml:"DeleteDot1XConfiguration"`
	Dot1XConfigurationToken onvif.ReferenceToken `xml:"Dot1XConfigurationToken"`
}

type DeleteDot1XConfigurationResponse struct {
}

type GetDot11Capabilities struct {
	XMLName string `xml:"GetDot11Capabilities"`
}

type GetDot11CapabilitiesResponse struct {
	XMLName string `xml:"GetDot11CapabilitiesResponse"`
	Capabilities onvif.Dot11Capabilities
}

type GetDot11Status struct {
	XMLName        string               `xml:"GetDot11Status"`
	InterfaceToken onvif.ReferenceToken `xml:"InterfaceToken"`
}

type GetDot11StatusResponse struct {
	XMLName string `xml:"GetDot11StatusResponse"`
	Status onvif.Dot11Status
}

type ScanAvailableDot11Networks struct {
	XMLName        string               `xml:"ScanAvailableDot11Networks"`
	InterfaceToken onvif.ReferenceToken `xml:"InterfaceToken"`
}

type ScanAvailableDot11NetworksResponse struct {
	Networks onvif.Dot11AvailableNetworks
}

type GetSystemUris struct {
	XMLName string `xml:"GetSystemUris"`
}

type GetSystemUrisResponse struct {
	XMLName string `xml:"GetSystemUrisResponse"`
	SystemLogUris   onvif.SystemLogUriList
	SupportInfoUri  xsd.AnyURI
	SystemBackupUri xsd.AnyURI
	Extension       xsd.AnyType
}

type StartFirmwareUpgrade struct {
	XMLName string `xml:"StartFirmwareUpgrade"`
}

type StartFirmwareUpgradeResponse struct {
	UploadUri        xsd.AnyURI
	UploadDelay      xsd.Duration
	ExpectedDownTime xsd.Duration
}

type StartSystemRestore struct {
	XMLName string `xml:"StartSystemRestore"`
}

type StartSystemRestoreResponse struct {
	UploadUri        xsd.AnyURI
	ExpectedDownTime xsd.Duration
}

type GetStorageConfigurations struct {
	XMLName string `xml:"GetStorageConfigurations"`
}

type GetStorageConfigurationsResponse struct {
	XMLName string `xml:"GetStorageConfigurationsResponse"`
	StorageConfigurations StorageConfiguration
}

type CreateStorageConfiguration struct {
	XMLName              string `xml:"CreateStorageConfiguration"`
	StorageConfiguration StorageConfigurationData
}

type CreateStorageConfigurationResponse struct {
	Token onvif.ReferenceToken
}

type GetStorageConfiguration struct {
	XMLName string               `xml:"GetStorageConfiguration"`
	Token   onvif.ReferenceToken `xml:"Token"`
}

type GetStorageConfigurationResponse struct {
	XMLName string `xml:"GetStorageConfigurationResponse"`
	StorageConfiguration StorageConfiguration
}

type SetStorageConfiguration struct {
	XMLName              string               `xml:"SetStorageConfiguration"`
	StorageConfiguration StorageConfiguration `xml:"StorageConfiguration"`
}

type SetStorageConfigurationResponse struct {
}

type DeleteStorageConfiguration struct {
	XMLName string               `xml:"DeleteStorageConfiguration"`
	Token   onvif.ReferenceToken `xml:"Token"`
}

type DeleteStorageConfigurationResponse struct {
}

type GetGeoLocation struct {
	XMLName string `xml:"GetGeoLocation"`
}

type GetGeoLocationResponse struct {
	XMLName string `xml:"GetGeoLocationResponse"`
	Location onvif.LocationEntity
}

//TODO: one or more Location
type SetGeoLocation struct {
	XMLName  string               `xml:"SetGeoLocation"`
	Location onvif.LocationEntity `xml:"Location"`
}

type SetGeoLocationResponse struct {
}

type DeleteGeoLocation struct {
	XMLName  string               `xml:"DeleteGeoLocation"`
	Location onvif.LocationEntity `xml:"Location"`
}

type DeleteGeoLocationResponse struct {
}
