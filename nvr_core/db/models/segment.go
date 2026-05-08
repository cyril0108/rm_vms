package models

// Segment represents a recorded 1-minute video file.
type Segment struct {
	ID        int64  `json:"id"`
	CameraID  int64  `json:"camera_id"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	FilePath  string `json:"file_path"`
	SizeBytes int64  `json:"size_bytes"`
}