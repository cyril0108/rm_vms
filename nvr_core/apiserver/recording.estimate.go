package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"nvr_core/apiserver/dto"
	"nvr_core/hardware"
	"nvr_core/utils"
	"sync"
	"time"
)

// ActiveStream represents a camera feed currently running on the NVR
type ActiveStream struct {
	CamID   int
	Profile string
	RTSPUrl string
}

func (api *APIServer) HandleGetRecordingEstimation(w http.ResponseWriter, r *http.Request) {

	ll := LOG.Prefix("[HandleGetRecordingEstimation]")

	utils.DisableHTTPTimeouts(w)

	doStorage := true
	nostorage := r.URL.Query().Get("nostorage")
	if nostorage != "" {
		doStorage = false
	}

	cams := api.PM.AllCameras()

	ctx := api.Context
	estimates := &dto.RecordingEstimates{}
	mapping := make(map[int]*dto.CameraRecordingEstimates, len(cams))

	var streams []*ActiveStream

	for _, cam := range cams {

		if cam.Active {

			sp := cam.GetProfile(utils.SegmentMainProfile)
			if sp != nil {
				streams = append(streams, &ActiveStream{
					CamID: cam.ID,
					Profile: utils.SegmentMainProfile,
					RTSPUrl: sp.Source,
				})
			}

			sp = cam.GetProfile(utils.SegmentSubProfile)
			if sp != nil {
				streams = append(streams, &ActiveStream{
					CamID: cam.ID,
					Profile: utils.SegmentSubProfile,
					RTSPUrl: sp.Source,
				})
			}

			mapping[cam.ID] = &dto.CameraRecordingEstimates{
				ID: cam.ID,
				Name: cam.Name,
			}

			if doStorage {

				bytes, err := api.Services.Camera.GetStorageSizeByCamera(ctx, int64(cam.ID))
				if err != nil {
					ll.Info("failed to get camera storage used", "cam", cam.ID, "err", err)
				}
				mapping[cam.ID].MBUsed = float64(bytes / 1000000)

			}

		}

	}


	estMB, err := api.calculateBandwidth(streams, mapping)
	if err != nil {

		utils.RespondJSONHTTPStatus(w, "Failed to calculate bandwidth", http.StatusInternalServerError)
		return

	}

	estimates.Cameras = utils.CopyMapValuesNL(mapping)

	estimates.MBPerMinute = estMB

	disk, err := hardware.GetDiskUsage(api.CFG.Server.StoragePath)
	if err != nil {

		utils.RespondJSONHTTPStatus(w, "Failed to get disk usage", http.StatusInternalServerError)
		return

	}

	availMB := float64(disk.AvailableBytes) / (1024*1024)
	availMB = availMB * utils.LowWaterMark
	estimates.AvailableMB = availMB

	if estMB > 0 {
		estimates.RecordingTime = availMB / estMB
	} else {
		estimates.RecordingTime = 0.0 
	}

	utils.RespondJSON(w, estimates, "success")

}

// CalculateTotalBandwidth acts as your API Handler logic
func (api *APIServer) calculateBandwidth(streams []*ActiveStream, estimates map[int]*dto.CameraRecordingEstimates) (float64, error) {
	var wg sync.WaitGroup

	// Mutex to protect totalMB while multiple goroutines add to it concurrently
	var mu sync.Mutex 
	var totalMB float64

	// Fan-Out: Launch a goroutine for every active stream simultaneously
	for _, stream := range streams {
		wg.Add(1)

		go func(s *ActiveStream) {
			defer wg.Done()

			// Create a strict 10-second timeout per probe.
			// If a camera is offline and C++ hangs, this prevents the Go API from deadlocking.
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// Send the IPC request to the estimation worker
			estimatedMB, err := api.PM.EstWorker.RequestProbe(ctx, s.CamID, s.Profile, s.RTSPUrl)

			if err != nil {
				fmt.Printf("[API] Failed to probe Cam %d %s: %v\n", s.CamID, s.Profile, err)
				return // Skip adding to the total
			}

			camE := estimates[stream.CamID]
			camE.Mbps = estimatedMB

			// Safely add the result to the total
			mu.Lock()
			totalMB += estimatedMB
			mu.Unlock()

		}(stream)
	}

	// Fan-In: Block here until every single goroutine has called wg.Done()
	wg.Wait()

	fmt.Printf("Final Calculation: All active streams consume %.2f MB/min\n", totalMB)
	return totalMB, nil
}