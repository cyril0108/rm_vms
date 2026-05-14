package discovery

import (
	"fmt"
	"net"
	"strings"
	"time"
)


const (
	wsDiscoveryPort  = 3702
	standardRtspPort = 554
)

var commonOnvifPorts = []int{80, 8080, 8899, 8000}

// VerifyResult holds the findings of the camera verification probe.
type VerifyResult struct {
	IP        string  `json:"ip"`
	IsValid   bool    `json:"is_valid"`
	InSystem  bool    `json:"in_system"`
	Protocol  string  `json:"protocol"`
	PortFound int     `json:"port_found"`
	RawData   string  `json:"raw_data"`
}

// Config allows customization of the verifier's behavior.
type Config struct {
	Timeout time.Duration
}

// Verifier handles targeted checks against specific IP addresses.
// Using a struct allows for dependency injection and state management 
// in the management layer.
type Verifier struct {
	timeout time.Duration
}

// New returns a customized Verifier instance.
func NewVerifier(cfg Config) *Verifier {
	if cfg.Timeout == 0 {
		cfg.Timeout = 2 * time.Second // Default fallback
	}
	return &Verifier{
		timeout: cfg.Timeout,
	}
}

// Verify inspects the target IP using Unicast WS-Discovery and TCP port sweeps.
func (v *Verifier) Verify(ip string) VerifyResult {
	// Try Unicast ONVIF Probe (Highest Confidence)
	if isONVIF, response := v.unicastProbe(ip); isONVIF {
		return VerifyResult{
			IP: ip,
			IsValid:   true,
			// Protocol:  "onvif-verified",
			Protocol:  "onvif",
			PortFound: wsDiscoveryPort,
			RawData:   response,
		}
	}

	// Fallback: Check standard RTSP port
	if v.checkPort(ip, standardRtspPort) {
		return VerifyResult{
			IP: ip,
			IsValid:   true,
			Protocol:  "rtsp",
			PortFound: standardRtspPort,
		}
	}

	// Comment our to speed up single ip scan
	// Fallback: Check common HTTP/ONVIF management ports
	// for _, port := range commonOnvifPorts {
	// 	if v.checkPort(ip, port) {
	// 		return VerifyResult{
	// 			IP: ip,
	// 			IsValid:   true,
	// 			Protocol:  "onvif-port-only",
	// 			PortFound: port,
	// 		}
	// 	}
	// }

	return VerifyResult{
		IsValid:  false,
		Protocol: "unknown",
	}
}

// checkPort attempts a raw TCP connection to determine if a service is listening.
func (v *Verifier) checkPort(ip string, port int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)

	conn, err := net.DialTimeout("tcp", address, v.timeout)
	if err != nil {
		return false
	}

	if conn != nil {
		conn.Close() // Immediately close to prevent socket exhaustion on the camera
		return true
	}
	return false
}

// unicastProbe sends a directed UDP WS-Discovery packet.
func (v *Verifier) unicastProbe(ip string) (bool, string) {
	address := fmt.Sprintf("%s:%d", ip, wsDiscoveryPort)

	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		return false, ""
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	payload := BuildProbeMessage()
	_, err = conn.Write([]byte(payload))
	if err != nil {
		return false, ""
	}

	// Apply the struct's timeout to the read operation
	conn.SetReadDeadline(time.Now().Add(v.timeout))

	buffer := make([]byte, 4096)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return false, ""
	}

	responseStr := string(buffer[:n])

	if strings.Contains(responseStr, "ProbeMatches") {
		return true, responseStr
	}

	return false, ""
}
