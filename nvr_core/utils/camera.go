package utils

import (
	"crypto/sha256"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// nvrNamespace is a custom, static UUIDv4 used as the mathematical root for our NVR.
// DO NOT change this once your system is in production, or all future IDs will shift.
var nvrNamespace = uuid.MustParse("7e297cc6-32fa-4b98-a1bf-ab0a822b7b9b")

// Generate creates a deterministic UUIDv5 based on hardware MAC or RTSP URL.
func Generate(macAddress string, rtspURL string) string {
	// Primary Strategy: Use the Hardware MAC Address (ONVIF Cameras)
	if macAddress != "" {
		cleanMAC := normalizeMAC(macAddress)
		// uuid.NewHash uses SHA256 under the hood when '5' (UUIDv5) is passed
		id := uuid.NewHash(sha256.New(), nvrNamespace, []byte(cleanMAC), 5)
		return id.String()
	}

	// Fallback Strategy: Use the sanitized RTSP URL (Manual/RTSP-only Cameras)
	if rtspURL != "" {
		cleanURL := sanitizeURL(rtspURL)
		id := uuid.NewHash(sha256.New(), nvrNamespace, []byte(cleanURL), 5)
		return id.String()
	}

	// Last Resort: If somehow both are empty, return a random UUIDv4 to prevent DB crashes
	return uuid.New().String()
}

// normalizeMAC strips colons/dashes and forces uppercase (e.g., "aa:bb:cc" -> "AABBCC")
func normalizeMAC(mac string) string {
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	return strings.ToUpper(strings.TrimSpace(mac))
}

// sanitizeURL removes user credentials and normalizes the scheme/host to lowercase
func sanitizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// If standard parsing fails, just trim whitespace and lowercase as a best-effort fallback
		return strings.ToLower(strings.TrimSpace(rawURL))
	}

	// CRITICAL: Strip the username and password so password changes don't alter the Camera ID
	parsed.User = nil

	// Ensure the host and scheme are lowercase (RTSP URLs are case-insensitive here)
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	// Return the clean URL (e.g., "rtsp://192.168.1.50/stream")
	return parsed.String()
}