package dto

import (
	"nvr_core/utils"
	"nvr_core/db/models"
)

type SegmentItem struct {
	ID         int    `json:"id,omitempty"`
	CameraID   int64  `json:"camera_id"`
	TimeRange
	DurationMs int64  `json:"duration_ms"`
	StreamURL  string `json:"stream_url"`
}

func NewSegmentItemFrom(segment *models.Segment) (*SegmentItem) {

	tr := TimeRange{
		StartTime: segment.StartTime,
		EndTime: segment.EndTime,
	}
	tr.ConvertToSeconds()

	return &SegmentItem {
		CameraID: segment.CameraID,
		TimeRange: tr,
		DurationMs: (segment.EndTime - segment.StartTime),
		StreamURL: utils.PathForCameraPlayURL(segment.CameraID, tr.StartTime),
	}

}

type SegmentSnapshot struct {
	ID          int    `json:"id,omitempty"`
	CameraID    int64  `json:"camera_id"`
	TimeRange
	SnapshotURL string `json:"url"`
}

func NewSegmentSnapshotFrom(segment *models.Segment) (*SegmentSnapshot) {

	tr := TimeRange{
		StartTime: segment.StartTime,
		EndTime: segment.EndTime,
	}
	tr.ConvertToSeconds()

	return &SegmentSnapshot {
		CameraID: segment.CameraID,
		TimeRange: tr,
		SnapshotURL: utils.PathForPlaybackSnapshotURL(segment.CameraID, tr.StartTime),
	}

}
