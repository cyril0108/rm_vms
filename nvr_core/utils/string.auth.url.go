package utils

import "net/url"

// InjectCredentials takes a raw RTSP URL and adds standard auth
// FFMpeg will not fees user/password directly. It will strip
// them from url, and feed them in case of need.
func InjectCredentials(rawURL, username, password string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Safely injects and URL-encodes the credentials
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String(), nil
}