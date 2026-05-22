package models

import "time"

const SegmentMainProfile = "main"
const SegmentSubProfile = "sub"

// Segment represents a recorded 1-minute video file.
type Segment struct {
	ID           int64   `json:"id"`
	Profile      string  `json:"profile"`
	CameraID     int64   `json:"camera_id"`
	StartTime    int64   `json:"start_time"`
	EndTime      int64   `json:"end_time"`
	FilePath     string  `json:"file_path"`
	SnapshotPath string  `json:"snapshot_path"`
	SizeBytes    int64   `json:"size_bytes"`
}

func (seg *Segment) StartTimeTime() time.Time {
	return time.UnixMilli(seg.StartTime)
}

func (seg *Segment) EndTimeTime() time.Time {
	return time.UnixMilli(seg.EndTime)
}

func (seg *Segment) IsSubProfile() bool {
	return seg.Profile == SegmentSubProfile
}