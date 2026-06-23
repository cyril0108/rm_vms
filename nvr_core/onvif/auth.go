package onvif

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"

	// "net/http/httputil"
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
	// Inject HTTP Basic Auth (Notice we allow empty password)
	// Many cameras require the username even if the password is blank.
	if ai.Username != "" {
		req.SetBasicAuth(ai.Username, ai.Password)
	}

	// Intercept and parse the outgoing XML body
	if req.Body != nil && ai.Username != "" {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			bodyString := string(bodyBytes)

			// Always generate our mathematically correct security block
			wsseHeader := generateWSSecurity(ai.Username, ai.Password)
			newHeader := fmt.Sprintf("<soap-env:Header>%s</soap-env:Header>", wsseHeader)

			// Check if the library generated an empty self-closing header
			if strings.Contains(bodyString, "<soap-env:Header/>") {
				bodyString = strings.Replace(bodyString, "<soap-env:Header/>", newHeader, 1)
			} else {
				// The library generated a broken header. Violently overwrite it.
				// This regex grabs everything from <soap-env:Header> to </soap-env:Header> 
				// and replaces the entire block.
				re := regexp.MustCompile(`(?s)<soap-env:Header>.*?</soap-env:Header>`)
				bodyString = re.ReplaceAllString(bodyString, newHeader)
			}

			// Repackage the request body and update the Content-Length
			req.Body = io.NopCloser(bytes.NewBuffer([]byte(bodyString)))
			req.ContentLength = int64(len(bodyString))
		}
	}

	// Send the request
	return ai.Proxied.RoundTrip(req)

	// Dump the outgoing request AFTER we've modified the XML
		// reqDump, err := httputil.DumpRequestOut(req, true)
		// if err == nil {
		// 	fmt.Println("\n========== RAW OUTGOING REQUEST ==========")
		// 	fmt.Println(string(reqDump))
		// 	fmt.Println("==========================================")
		// }

		// // Send the request
		// resp, err := ai.Proxied.RoundTrip(req)
		// if err != nil {
		// 	return resp, err
		// }

		// // Dump the raw incoming response from the camera
		// respDump, err := httputil.DumpResponse(resp, true)
		// if err == nil {
		// 	fmt.Println("\n========== RAW INCOMING RESPONSE ==========")
		// 	fmt.Println(string(respDump))
		// 	fmt.Println("===========================================\n")
		// }

		// return resp, err
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