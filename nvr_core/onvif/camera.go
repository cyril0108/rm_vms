package onvif

// DiscoveredCamera represents a network video transmitter found on the LAN.
type DiscoveredCamera struct {
	MessageID string
	Address   string // Unique endpoint reference
	Types     string // ONVIF device types
	Scopes    string // Location, hardware, name
	XAddrs    string // Space-separated service URLs used for further ONVIF requests
}


// OnvifRecord
type OnvifRecord struct {
	IP           string  `json:"ip"`
	MACAddress   string  `json:"mac_address"`
	Manufacturer string  `json:"manufacturer"`
	Model        string  `json:"model"`
	Firmware     string  `json:"firmware"`
	SerialNumber string  `json:"serial_number"`
	SupportsPTZ  bool    `json:"supports_ptz"`

	// Streams
	MainStream       string  `json:"mainstream"`
	SubStream        string  `json:"substream"`
	MainStreamToken  string  `json:"mainstream_token"`
	SubStreamToken   string  `json:"substream_token"`

	// Camera user/pwd
	Username         string  `json:"username"`
	Password         string  `json:"password"`

	ErrorMSG         string  `json:"error_msg"`
}

// BulkScanRequest represents the JSON payload from the frontend
type BulkScanRequest struct {
	StartIP  string `json:"start_ip"`
	EndIP    string `json:"end_ip"`   // Optional: if empty, scans primary subnet
	Username string `json:"username"`
	Password string `json:"password"`
}
