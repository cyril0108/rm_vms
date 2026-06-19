package apiserver

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"nvr_core/network"
	"nvr_core/onvif"
	"nvr_core/onvif/discovery"
	"nvr_core/utils"
)

// HandleBulkONVIFScan performs a fast sweep followed by authenticated data fetching
func (s *APIServer) HandleBulkONVIFScan(w http.ResponseWriter, r *http.Request) {
	var req onvif.BulkScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondErrInvalidPayload(w)
		return
	}

	ctx := r.Context()
	var sweepResults []discovery.VerifyResult

	// ==========================================
	// PHASE 1: The Fast UDP Sweep
	// ==========================================
	log.Println("[BulkScan] Starting Phase 1: Fast UDP Sweep...")

	if req.StartIP == "" {
		baseIP, err := network.GetPrimarySubnetBase()
		if err != nil {
			utils.RespondJSONHTTPStatus(w, "Failed to detect LAN subnet", http.StatusInternalServerError)
			return
		}
		sweepResults = ScanSweepPrimarySubnet(ctx, baseIP)
	} else {
		if err := network.IsSafeTarget(req.StartIP); err != nil {
			utils.RespondJSONHTTPStatus(w, err.Error(), http.StatusBadRequest)
			return
		}
		sweepResults, _ = ScanSweepSubnetRange(ctx, req.StartIP, req.EndIP)
	}

	if len(sweepResults) == 0 {
		utils.RespondJSON(w, []onvif.OnvifRecord{}, "") // Return empty array if no cameras found
		return
	}

	log.Printf("[BulkScan] Phase 1 Complete. Found %d potential ONVIF devices. Starting Phase 2...", len(sweepResults))

	log.Printf("[BulkScan] user %s p %s", req.Username, req.Password)

	// ==========================================
	// PHASE 2: Concurrent Data Fetching
	// ==========================================
	var detailedRecords []*onvif.OnvifRecord
	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex is required to safely append to the slice from multiple Goroutines

	// Spin up a Goroutine for each confirmed camera
	for _, cam := range sweepResults {
		if cam.Protocol == "onvif" || cam.Protocol == "onvif-verified" {
			wg.Add(1)
			
			go func(targetIP string) {
				defer wg.Done()

				// Only fetch heavy data for IPs we KNOW are cameras
				record, err := onvif.FetchCameraONVIFData(targetIP, DefaultScanPort, req.Username, req.Password)
				if err != nil {
					log.Printf("[BulkScan] Failed to authenticate or fetch data for %s: %v", targetIP, err)
					return // Skip this camera, likely wrong password
				}

				// Safely lock, append, and unlock
				mu.Lock()
				detailedRecords = append(detailedRecords, record)
				mu.Unlock()

			}(cam.IP)
		}
	}

	// Wait for all HTTP fetch requests to finish
	wg.Wait()

	log.Printf("[BulkScan] Phase 2 Complete. Successfully fetched details for %d cameras.", len(detailedRecords))

	// Return the detailed records to the frontend
	if detailedRecords == nil {
		detailedRecords = make([]*onvif.OnvifRecord, 0)
	}

	detailedRecords = s.ApplyCamerasOnvifRecordInSystemCheck(ctx, detailedRecords)

	if err := utils.RespondJSON(w, detailedRecords, ""); err != nil {
		log.Printf("Error encoding bulk results: %v", err)
	}
}