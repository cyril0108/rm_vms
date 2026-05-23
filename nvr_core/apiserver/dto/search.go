package dto

// TimelineBlock represents a continuous block of recorded video.
type TimelineBlock struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type SearchResponse struct {
	CameraID     int           `json:"camera_id"`
	SearchWindow struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	} `json:"search_window"`
	Segments []SegmentItem `json:"segments"`
}

type TimelineResponse struct {
	CameraID   int              `json:"camera_id"`
	Timelines  []TimelineBlock  `json:"timelines"`
}

func (tb *TimelineBlock) ConvertToSeconds() {
	tb.StartTime = tb.StartTime / 1000
	tb.EndTime = tb.EndTime / 1000
}