package network

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)


func GetPrimaryIP() (string, error) {

	// 8.8.8.8 is used as a dummy target. No actual traffic is sent.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("could not determine local IP: %w", err)
	}
	defer conn.Close()

	// Get the local IP address
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil

}


// GetPrimaryNIC reads the Linux routing table to find the interface 
// handling the default gateway (0.0.0.0).
func GetPrimaryNIC() string {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "eth0" // Safe fallback if /proc isn't available
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// The /proc/net/route file is tab/space separated.
		// Column 0 is Iface, Column 1 is Destination.
		fields := strings.Fields(scanner.Text())

		// A default route has a destination of 00000000 in hex
		if len(fields) >= 2 && fields[1] == "00000000" {
			return fields[0] // e.g., returns "enp3s0" or "eth0"
		}
	}

	// Fallback if no default route is found (e.g., completely offline system)
	return "eth0" 
}

// GetAvailableNICs returns a list of real physical interfaces for the Vue dropdown
func GetAvailableNICs() []string {
	ll := LOG.Prefix("[GetAvailableNICs]")
	var validNICs []string
	interfaces, err := net.Interfaces()
	if err != nil {
		ll.Error("cannot get interfaces", "error", err)
		return validNICs
	}

	for _, iface := range interfaces {
		// Ignore loopback and interfaces that are powered down
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		// Ignore common Docker/Virtual network prefixes
		if strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "veth") || strings.HasPrefix(iface.Name, "br-") {
			continue
		}

		validNICs = append(validNICs, iface.Name)
	}
	return validNICs
}