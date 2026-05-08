package dto

import (
	"nvr_core/utils"
	"nvr_core/db/models"
)

type SegmentItem struct {
	ID         int    `json:"id"`
	CameraID   int64  `json:"camera_id"`
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
	DurationMs int64  `json:"duration_ms"`
	StreamURL  string `json:"stream_url"`
}

func NewSegmentItemFrom(segment *models.Segment) (SegmentItem) {

	return SegmentItem {
		CameraID: segment.CameraID,
		StartTime: segment.StartTime,
		EndTime: segment.EndTime,
		DurationMs: (segment.EndTime - segment.StartTime),
		StreamURL: utils.PathForCameraPlayURL(segment.CameraID, segment.StartTime),
	}

}