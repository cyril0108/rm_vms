package apiserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"nvr_core/network"
	"nvr_core/onvif"
	"nvr_core/onvif/discovery"
	"time"
)

// GetCameras safely iterates over the sync.Map
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
	log.Printf("[HandleCameraScan] Start scan process")

	// Detect the subnet
	baseIP, err := network.GetPrimarySubnetBase()
	if err != nil {
		http.Error(w, "Failed to detect LAN subnet", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[HandleCameraScan] Sweeping subnet: %s0/24", baseIP)

	// Initialize the Verifier (from the Unicast code)
	v := discovery.NewVerifier(discovery.Config{
		Timeout: 2 * time.Second, // Keep it short for a fast sweep
	})

	// Execute the concurrent sweep
	ctx := r.Context()
	result := v.SweepSubnet(ctx, baseIP)

	if(result == nil) {
		result = make([]discovery.VerifyResult, 0)
	}

	// Return results
	if err := RespondJSON(w, result); err != nil {
		log.Printf("Error encoding results: %v", err)
	}
}

// HandleFetchCameraONVIF
func (s *APIServer) HandleFetchCameraONVIF(w http.ResponseWriter, r *http.Request) {

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
