package onvif

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
)


func ONVIFAddress(ip string, port int) string {
	return fmt.Sprintf("%s:%d", ip, port)
}


// xmlDumpInterceptor catches and logs the raw SOAP XML
type xmlDumpInterceptor struct {
	Proxied http.RoundTripper
}

// Compile the regex globally so it doesn't slow down HTTP requests.
// This targets the opening (<) and closing (</) tags of standard ONVIF namespaces.
const PTZNamespaceReplace = `xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"`

var (
	onvifRespNamespaceRegex = regexp.MustCompile(`(</?)(tt|trt|tptz|tds|tev|timg|tmd|tan|tr2|wsnt):`)

	tptzRegex = regexp.MustCompile(`xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"`)
	// Matches any <PanTilt ...> or <ns:PanTilt ...>
	panTiltRegex = regexp.MustCompile(`<(/[a-zA-Z0-9_-]+:)?PanTilt`)
	// Matches any <Zoom ...> or <ns:Zoom ...>
	zoomRegex    = regexp.MustCompile(`<(/[a-zA-Z0-9_-]+:)?Zoom`)

	boundsRegex = regexp.MustCompile(`<(/)?([a-zA-Z0-9_-]+:)?Bounds`)

)

func (x *xmlDumpInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {

	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {

			// --- THE NAMESPACE SANITIZER ---

			bodyBytes = bytes.Replace(bodyBytes, []byte("<Envelope"), []byte("<Envelope xmlns:tt=\"http://www.onvif.org/ver10/schema\""), 1)
			// bodyBytes = tptzRegex.ReplaceAll(bodyBytes, []byte("${1} "+PTZNamespaceReplace))
			// Forcefully rewrite <PanTilt> to <tt:PanTilt>
			bodyBytes = panTiltRegex.ReplaceAll(bodyBytes, []byte("<${1}tt:PanTilt"))
			// Forcefully rewrite <Zoom> to <tt:Zoom>
			bodyBytes = zoomRegex.ReplaceAll(bodyBytes, []byte("<${1}tt:Zoom"))

			// Log the SANITIZED payload
			fmt.Printf("\n[ONVIF DEBUG] --- OUTGOING REQUEST TO %s ---\n%s\n------------------------------------------\n", req.URL.String(), string(bodyBytes))

			// Re-pack the modified bytes into the request
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			req.ContentLength = int64(len(bodyBytes)) // Update the length!
		}
	}

	// Intercept and Dump the Request
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			// Print the raw XML Request
			fmt.Printf("\n[ONVIF DEBUG] --- OUTGOING REQUEST TO %s ---\n%s\n------------------------------------------\n", req.URL.String(), string(bodyBytes))

			// RESTORE THE BODY so the actual HTTP client can send it
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// Execute the actual network request
	res, err := x.Proxied.RoundTrip(req)
	if err != nil {
		return res, err
	}

	// (Optional) Intercept and Dump the Response from the Camera
	if res.Body != nil {
		resBytes, err := io.ReadAll(res.Body)
		if err == nil {

			fmt.Printf("\n[ONVIF DEBUG] --- INCOMING RESPONSE ---\n%s\n---------------------------------------\n", string(resBytes))

			// RESTORE THE BODY so the onvif-go parser can read it
			res.Body = io.NopCloser(bytes.NewBuffer(resBytes))
		}
	}

	return res, nil
}