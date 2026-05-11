package network

import (
	"encoding/binary"
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

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address format: %s", ipStr)
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", fmt.Errorf("address is not IPv4: %s", ipStr)
	}

	baseIP := fmt.Sprintf("%d.%d.%d.", ipv4[0], ipv4[1], ipv4[2])
	return baseIP, nil
}

func GetSubnetParts(ipStr string) (string, int, error) {
	// ParseIP strictly validates the format. 
	// It returns nil if the string is not a valid IP address.
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", -1, fmt.Errorf("invalid IP address format: %s", ipStr)
	}

	// Ensure it is specifically an IPv4 address.
	// (ParseIP handles IPv6 too, so we must filter them out).
	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", -1, fmt.Errorf("address is not IPv4: %s", ipStr)
	}

	// Extract directly from the raw byte array!
	// ipv4[0], ipv4[1], ipv4[2], ipv4[3] map perfectly to the 4 octets.
	baseIP := fmt.Sprintf("%d.%d.%d.", ipv4[0], ipv4[1], ipv4[2])
	adr := int(ipv4[3])

	return baseIP, adr, nil
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


/**
 * ==========================================
 * Advanced Subnet Methods that may be useful
 * in larger network scale.
 * ==========================================
 */

// GetSubnetIPs takes a CIDR string (e.g., "192.168.1.50/22") 
// and returns a list of every usable IP address in that subnet for probing.
func GetSubnetIPs(cidrStr string) ([]net.IP, error) {

	// Parse the CIDR string.
	// ip is the specific device IP (192.168.1.50)
	// ipNet contains the calculated Network Base and Mask.
	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR format: %w", err)
	}

	// Ensure it is an IPv4 network
	ipv4 := ipNet.IP.To4()
	if ipv4 == nil {
		return nil, fmt.Errorf("address is not IPv4: %s", cidrStr)
	}

	// Convert the Network Base IP to a 32-bit integer for fast math
	networkIP := binary.BigEndian.Uint32(ipv4)

	// Calculate the subnet size based on the mask
	mask := binary.BigEndian.Uint32(ipNet.Mask)

	// The bitwise NOT (^) of the mask gives us the total number of IPs in the block
	// e.g., for a /24, ^mask == 255.
	hostCount := ^mask

	// Edge case: /32 networks only have 1 IP, no broadcast or network address
	if hostCount == 0 {
		return []net.IP{ipNet.IP}, nil
	}

	// Pre-allocate the slice to prevent memory reallocation
	// We subtract 2 to exclude the Network Address (.0) and Broadcast Address (.255)
	usableIPs := make([]net.IP, 0, hostCount-1)

	// Generate every IP using integer addition
	for i := uint32(1); i < hostCount; i++ {
		currentIPInt := networkIP + i
		
		// Convert the 32-bit integer back into a 4-byte net.IP array
		currentIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(currentIP, currentIPInt)
		
		usableIPs = append(usableIPs, currentIP)
	}

	return usableIPs, nil
}

// GetActiveSubnets scans all physical interfaces and returns the active IPv4 networks.
func GetActiveSubnets() ([]*net.IPNet, error) {
	var activeNetworks []*net.IPNet

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, address := range addrs {
		// Use type assertion to ensure it's a network address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			
			// Strictly filter for IPv4
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				
				// Optional: Skip known Docker bridges (172.17.x.x)
				// Using raw bytes instead of strings!
				if ipv4[0] == 172 && ipv4[1] == 17 {
					continue 
				}

				activeNetworks = append(activeNetworks, ipnet)
			}
		}
	}

	return activeNetworks, nil
}