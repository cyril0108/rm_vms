package dto

// ===========================
// Recording Estimates
// ===========================
type RecordingEstimates struct {
	MBPerMinute    float64  `json:"mb_per_min"`
	AvailableMB    float64  `json:"mb_available"`
	RecordingTime  float64  `json:"recording_time"`
	Cameras        []*CameraRecordingEstimates  `json:"cameras,omitempty"`
}

type CameraRecordingEstimates struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	MBUsed         float64  `json:"mb_used"`
	Mbps           float64  `json:"mbps"`
	RecordedTime   float64  `json:"recorded_min"`
	RecordingTime  float64  `json:"recording_min"`
}


// Return the estimate recording time by the given mbps
// based on the AvailableMB
func (re *RecordingEstimates) CalculateRecordingTime(mbps float64) float64 {
	if mbps > 0 {
		return re.AvailableMB / mbps
	}
	return 0.0
}
