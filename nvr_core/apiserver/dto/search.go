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
	tb.StartTime = (tb.StartTime / 1000) + 1
	tb.EndTime = tb.EndTime / 1000
	// Explaine:
	// Assume start time 123456(ms) to end time 183779(ms)
	// We normally send 123(s) search.
	// * = 123*1000 = 123000
	// X = (123+1)*1000 = 124000
	// ---------*--|123456--X-----------|183779-----------
	// This shows that X is what we want.
	// 
	// The only risk is that when the recorded range is
	// within 1 or 2 seconds, then this method could fail.
	// On those extreme scenario, one should really
	// use mstime.
}