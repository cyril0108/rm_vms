package license

import (
	"context"
	"fmt"

	"nvr_core/reqapi/apiclient"
)

// Service is the dedicated client for License-related endpoints
type Service struct {
	client *apiclient.Client
}

// NewService instantiates the license API wrapper
func NewService(baseURL string) (*Service, error) {
	apiClient, err := apiclient.NewClient(baseURL)
	if err != nil {
		return nil, err
	}
	return &Service{client: apiClient}, nil
}


// ====================================================
// API calls
// ====================================================

// GetInfo maps to GET /license/{key}
func (s *Service) GetLicense(ctx context.Context, key string) (*LicenseAPIResponse, error) {
	path := fmt.Sprintf("/license/%s", key)
	
	var result LicenseAPIResponse
	// Notice how clean the call is: Method, Path, Body(nil), ResponseTarget
	err := s.client.Do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	
	return &result, nil
}

// ApplyNew maps to POST /license/apply
func (s *Service) Apply(ctx context.Context, kind, machineID string) error {

	// kind = "trial"

	reqData := ApplyRequest{
		Kind:       kind,
		MachineID: machineID,
	}
	
	// Passing nil for responseTarget since we only care if it succeeds (HTTP 200)
	return s.client.Do(ctx, "POST", "/license", reqData, nil)
}