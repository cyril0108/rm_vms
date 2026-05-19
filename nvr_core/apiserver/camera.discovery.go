package apiserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/network"
	"nvr_core/onvif"
	"nvr_core/onvif/discovery"
	"time"
)

// Check Single IP
func (s *APIServer) HandleCameraProbe(w http.ResponseWriter, r *http.Request) {

	v := discovery.NewVerifier(discovery.Config{
		Timeout: 3 * time.Second,
	})

	targetIP := r.PathValue("ip")
	if targetIP == "" {
		http.Error(w, "No IP", http.StatusBadRequest)
		return
	}

	result := v.Verify(targetIP)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error probing camera: %v", err)
	}

}

/**
 * ===========================================
 * Fetch ONVIF Detail Data
 * ===========================================
 */

// HandleFetchCameraONVIF godoc
// @Summary      Fetch ONVIF data
// @Description  Fetch ONVIF camera detail data
// @Tags         Cameras, Scan
// @Accept       json
// @Produce      json
// @Param        camera  body      dto.CreateCameraRequest  true  "ONVIF credentials and payload"
// @Success      201     {object}  onvif.OnvifRecord
// @Failure      400     {string}  string "Invalid JSON payload"
// @Failure      500     {string}  string "Failed to get ONVIF data or internal server error"
// @Router       /api/scan/{ip}/onvif [post]
func (s *APIServer) HandleFetchCameraONVIF(w http.ResponseWriter, r *http.Request) {

	targetIP := r.PathValue("ip")
	if targetIP == "" {
		http.Error(w, "No IP", http.StatusBadRequest)
		return
	}

	var cam dto.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&cam); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	result, err := onvif.FetchCameraONVIFData(targetIP, cam.Username, cam.Password)
	if result==nil && err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		result.ErrorMSG = err.Error()
	}

	if err := RespondJSON(w, result); err != nil {
		log.Printf("Error fetching camera ONVIF data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

/**
 * ===========================================
 * Scan Methods
 * ===========================================
 */

/**
 * Scan using multicast method
 */
func (s *APIServer) HandleCameraScan(w http.ResponseWriter, r *http.Request) {

	log.Printf("[HandleCameraScan] Start scan process")

	scanner, err := discovery.NewScanner()

	log.Printf("[HandleCameraScan] New scanner")

	scanTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(r.Context(), scanTimeout)
	defer cancel()

	log.Printf("[HandleCameraScan] begin scan with context")

	result, err := scanner.Scan(ctx)
	if err != nil {
		http.Error(w, "Error scannig cameras.", http.StatusBadGateway)
		return
	}

	if(result == nil) {
		result = make([]onvif.DiscoveredCamera, 0)
	}

	if err := RespondJSON(w, result); err != nil {
		log.Printf("Error probing camera: %v", err)
	}

}

/**
 * Sweep every devices of the Detected subnet
 */
func (s *APIServer) HandleCameraSweep(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HandleCameraSweep] Start scan sweep process")

	// Read the query parameters
	startIP := r.URL.Query().Get("startip")
	endIP := r.URL.Query().Get("endip")

	ctx := r.Context()
	var result []discovery.VerifyResult

	if startIP == "" && endIP == "" {

		// Default: scan primary subnet
		// Detect the subnet
		baseIP, err := network.GetPrimarySubnetBase()
		if err != nil {
			http.Error(w, "Failed to detect LAN subnet", http.StatusInternalServerError)
			return
		}

		if err := network.IsSafeTarget(baseIP+"1"); err != nil {
			http.Error(w, "Failed to get private subnet.", http.StatusInternalServerError)
			return
		}

		result = ScanSweepPrimarySubnet(ctx, baseIP)

	} else {

		// Prevent non-private ip can
		if err := network.IsSafeTarget(startIP); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := ScanSweepSubnetRange(ctx, startIP, endIP)
		if err != nil {
			http.Error(w, "Failed scan subnet range", http.StatusInternalServerError)
			return
		}

		result = res

	}

	if(result == nil) {
		result = make([]discovery.VerifyResult, 0)
	}

	result = s.ApplyCamerasInSystemCheck(ctx, result)

	// Return results
	if err := RespondJSON(w, result); err != nil {
		log.Printf("Error encoding results: %v", err)
	}
}

/**
 * ------------------------------
 * Private helper functions
 * ------------------------------
 */

func ScanSweepPrimarySubnet(ctx context.Context, baseIP string) []discovery.VerifyResult {

	// Initialize the Verifier (from the Unicast code)
	v := discovery.NewVerifier(discovery.Config{
		Timeout: 2 * time.Second, // Keep it short for a fast sweep
	})

	log.Printf("[HandleCameraScan] Sweeping subnet: %s0/24", baseIP)

	// Execute the concurrent sweep
	return v.SweepSubnet(ctx, baseIP)
}

func ScanSweepSubnetRange(ctx context.Context, startip string, endip string) ([]discovery.VerifyResult, error) {

	// Initialize the Verifier (from the Unicast code)
	v := discovery.NewVerifier(discovery.Config{
		Timeout: 2 * time.Second, // Keep it short for a fast sweep
	})

	// Execute the concurrent sweep
	return v.SweepSubnetIPRange(ctx, startip, endip)

}

