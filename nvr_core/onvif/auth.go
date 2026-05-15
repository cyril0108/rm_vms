package onvif

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AuthInterceptor ensures HTTP Basic Auth is attached and manually injects 
// WS-Security XML if the underlying library fails to do so (e.g., blank passwords).
type AuthInterceptor struct {
	Proxied  http.RoundTripper
	Username string
	Password string
}

func (ai *AuthInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	// 1. Inject HTTP Basic Auth (Notice we removed the Password != "" check!)
	// Many cameras require the username even if the password is blank.
	if ai.Username != "" {
		req.SetBasicAuth(ai.Username, ai.Password)
	}

	// 2. Intercept and parse the outgoing XML body
	if req.Body != nil && ai.Username != "" {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			bodyString := string(bodyBytes)

			// 3. If the library skipped adding the security header, inject it ourselves
			if !strings.Contains(bodyString, "wsse:Security") {
				wsseHeader := generateWSSecurity(ai.Username, ai.Password)

				// The go-use/onvif library outputs <soap-env:Header/> when empty.
				// We string-replace it with our populated header.
				bodyString = strings.Replace(
					bodyString, 
					"<soap-env:Header/>", 
					fmt.Sprintf("<soap-env:Header>%s</soap-env:Header>", wsseHeader), 
					1,
				)

				// Reassign the modified payload
				bodyBytes = []byte(bodyString)
			}

			// 4. Repackage the request body and CRITICALLY update the Content-Length
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			req.ContentLength = int64(len(bodyBytes))
		}
	}

	// Send the request
	return ai.Proxied.RoundTrip(req)
}

// generateWSSecurity creates the exact XML block needed for ONVIF authentication
func generateWSSecurity(username, password string) string {
	// Generate Nonce
	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	nonce64 := base64.StdEncoding.EncodeToString(nonceBytes)

	// Generate Timestamp
	created := time.Now().UTC().Format(time.RFC3339)

	// Calculate PasswordDigest
	h := sha1.New()
	h.Write(nonceBytes)
	h.Write([]byte(created))
	h.Write([]byte(password)) // This perfectly handles password == ""
	digest64 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Return the formatted XML string
	return fmt.Sprintf(`
		<wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
			<wsse:UsernameToken>
				<wsse:Username>%s</wsse:Username>
				<wsse:Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">%s</wsse:Password>
				<wsse:Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">%s</wsse:Nonce>
				<wsu:Created xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">%s</wsu:Created>
			</wsse:UsernameToken>
		</wsse:Security>`, username, digest64, nonce64, created)
}