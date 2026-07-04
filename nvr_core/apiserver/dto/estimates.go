package dto


// ===========================
// Recording Estimates
// ===========================
type RecordingEstimates struct {
	MBPerMinute    float64  `json:"mb_per_min"`
	AvailableMB    float64  `json:"mb_available"`
	RecordingTime  float64  `json:"recording_time"`
}


