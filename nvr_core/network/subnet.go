package network

import (
	"fmt"
	"net"
	"strings"
)

// GetPrimarySubnetBase uses the routing table to find the main LAN IP.
// Returns a string like "192.168.1."
func GetPrimarySubnetBase() (string, error) {

	ipStr, err := GetPrimaryIP()
	if err != nil {
		return "", err
	}

	// Split the IP (e.g., "192.168.1.50") and reconstruct the base ("192.168.1.")
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid IPv4 address format")
	}

	baseIP := fmt.Sprintf("%s.%s.%s.", parts[0], parts[1], parts[2])
	return baseIP, nil
}

// GetAllSubnetBases scans all physical interfaces for active IPv4 subnets.
func GetAllSubnetBases() ([]string, error) {
	var bases []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, address := range addrs {
		// Check if it's a valid IP network and NOT a loopback address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// We only care about IPv4 for this sweep
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()

				parts := strings.Split(ipStr, ".")
				if len(parts) == 4 {
					base := fmt.Sprintf("%s.%s.%s.", parts[0], parts[1], parts[2])

					// Optional: Filter out known Docker bridge subnets (e.g., 172.17.x.x) if needed
					if !strings.HasPrefix(base, "172.17.") {
						bases = append(bases, base)
					}
				}
			}
		}
	}

	return bases, nil
}