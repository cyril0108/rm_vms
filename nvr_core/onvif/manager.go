package onvif

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"nvr_core/db/models"
	onvifxml "nvr_core/onvif/xml"

	goonvif "github.com/use-go/onvif"
)

// ONVIFManager handles the core connection and capability probing.
// It owns the *onvif.Device but delegates specific actions (like PTZ) to other services.
type ONVIFManager struct {
	Device *goonvif.Device
	Camera *models.Camera
	stepMu       sync.Mutex
}

// NewONVIFManager creates the connection and automatically fetches basic capabilities.
func NewONVIFManager(cfg DeviceConfig, cam *models.Camera) (*ONVIFManager, error) {

	if cfg.Port == 0 {
		cfg.Port = 80 // Default ONVIF port
	}

	if cfg.HTTPClient == nil {

		// Inject the custom RoundTripper into the HTTP Client
		cfg.HTTPClient = &http.Client{
			Transport: &xmlDumpInterceptor{
				Proxied: http.DefaultTransport,
			},
			// Optional: You can set a timeout here as well
			Timeout: 10 * time.Second, 
		}


	}

	xaddr := ONVIFAddress(cfg.IP, cfg.Port)

	ll := LOG.Prefix("[NewONVIFManager]")
	ll.Info("", "xaddr", xaddr)

	params := goonvif.DeviceParams{
		Xaddr:      xaddr,
		Username:   cfg.Username,
		Password:   cfg.Password,
		HttpClient: cfg.HTTPClient, 
	}

	// NewDevice automatically makes the GetCapabilities / GetServices SOAP calls
	// under the hood and maps the supported namespaces.
	dev, err := goonvif.NewDevice(params)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize onvif device: %w", err)
	}

	// dev.GetDeviceInfo()

	return &ONVIFManager{
		Device: dev,
		Camera: cam,
	}, nil
}


// GetDeviceDetails fetches the physical hardware information from the camera.
func (m *ONVIFManager) GetDeviceInfo() (*goonvif.DeviceInfo) {
	info := m.Device.GetDeviceInfo()
	return &info

}


// ---------------------------------------------------------
// CAPABILITY DETECTION
// ---------------------------------------------------------

// SupportsPTZ checks if the camera exposed a PTZ service endpoint during handshake.
func (c *ONVIFManager) SupportsPTZ() bool {
	// use-go/onvif stores the discovered services internally.
	endpoint := c.Device.GetEndpoint("ptz")
	return endpoint != ""
}

// SupportsMedia checks if the camera can stream video and provides profiles.
func (c *ONVIFManager) SupportsMedia() bool {
	endpoint := c.Device.GetEndpoint("media")
	return endpoint != ""
}

// SupportsEvents checks if the camera supports motion/alarm event subscriptions.
func (c *ONVIFManager) SupportsEvents() bool {
	endpoint := c.Device.GetEndpoint("events")
	return endpoint != ""
}

func RangeExists(r onvifxml.FloatRange) bool {
	return r.Max != 0 || r.Min != 0
}

func CallONVIFRequest[T any](device *goonvif.Device, req interface{}) (T, error) {
	// Initialize an empty variable of type T to return in case of errors
	var emp T

	// Execute the network request
	res, err := device.CallMethod(req)
	if err != nil {
		return emp, fmt.Errorf("failed to call ONVIF method: %w", err)
	}

	// Ensure the body is closed after reading to prevent memory leaks
	defer res.Body.Close()

	// Read the raw response bytes
	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return emp, fmt.Errorf("failed to read response body: %w", err)
	}

	// Decode the generic envelope
	var env onvifxml.Envelope[T]
	if err := xml.NewDecoder(bytes.NewReader(resBytes)).Decode(&env); err != nil {
		return emp, fmt.Errorf("failed to parse XML response: %w", err)
	}

	// Return the successfully parsed inner content and a nil error
	return env.Body.Content, nil
}
