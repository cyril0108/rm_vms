package network

import (
	"fmt"
	"net"
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
