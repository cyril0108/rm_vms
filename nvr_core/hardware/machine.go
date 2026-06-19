package hardware

// For Linux:
// sudo cat /sys/class/dmi/id/product_uuid

// const fk_machine_uuid = "0596f0717bbfa9afb0c4d0109395eafcd1"

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"strings"
)

const UNKNOWN_HARDWARE_ID = "UNKNOWN_HARDWARE_ID"

// GetPersistentMachineID returns a deterministic SHA-256 hash of the physical hardware.
func GetPersistentMachineID() string {
	// Read hardware-burned identifiers from SMBIOS/DMI tables
	sysUUID := readSysFile("/sys/class/dmi/id/product_uuid")
	boardSerial := readSysFile("/sys/class/dmi/id/board_serial")

	// If DMI data is completely empty (can happen on cheap ARM SBCs like old Raspberry Pis),
	// fallback to the primary MAC address.
	if sysUUID == "" && boardSerial == "" {
		fallbackID := getPrimaryMACAddress()
		return hashID(fallbackID)
	}

	// Concatenate the hardware identifiers
	rawFingerprint := sysUUID + "::" + boardSerial

	return hashID(rawFingerprint)
}

// readSysFile safely reads and trims a file from the sysfs virtual filesystem.
func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "" // Return empty string if permission denied or file missing
	}
	// Hardware strings often contain trailing null bytes or newlines
	return strings.TrimRight(string(data), "\n\r\x00 ")
}

// hashID creates a uniform 64-character hex string from the raw hardware data.
func hashID(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// getPrimaryMACAddress is a fallback for systems lacking SMBIOS support.
func getPrimaryMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return UNKNOWN_HARDWARE_ID
	}

	for _, iface := range interfaces {
		// Skip loopback, virtual interfaces, and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		// Prefer standard physical ethernet naming conventions over virtual bridges (docker0, veth)
		if strings.HasPrefix(iface.Name, "eth") || strings.HasPrefix(iface.Name, "en") {
			return iface.HardwareAddr.String()
		}
	}

	// Ultimate fallback to whatever has a MAC
	for _, iface := range interfaces {
		if iface.HardwareAddr != nil && iface.HardwareAddr.String() != "" {
			return iface.HardwareAddr.String()
		}
	}
	
	return UNKNOWN_HARDWARE_ID
}