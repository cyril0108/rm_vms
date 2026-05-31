package m3u8

import (
	"fmt"
	"math"
	"strings"

	"nvr_core/db/models"
	"nvr_core/utils"
)

type M3U8Builder struct {
	camID   int64
	baseURL string
	builder strings.Builder
}

func NewM3U8Builder(id int64, url string) M3U8Builder {
	return M3U8Builder {
		camID: id,
		baseURL: url,
	}
}

func (b *M3U8Builder) NL() {
	b.builder.WriteString("\n")
}

func (b *M3U8Builder) Begin() {
	b.builder.WriteString("#EXTM3U\n")
}

func (b *M3U8Builder) XVOD() {
	b.builder.WriteString("#EXT-X-VERSION:3\n")
	b.builder.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")

}

func (b *M3U8Builder) XTargetDuration(dur int) {
	// Max segment duration
	b.builder.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", dur)) 
}

func (b *M3U8Builder) XMediaSequence() {
	b.builder.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")
}


func (b *M3U8Builder) XDiscontinuity() {
	b.builder.WriteString("#EXT-X-DISCONTINUITY\n")
}

func (b *M3U8Builder) XVODEnd() {
	b.builder.WriteString("#EXT-X-ENDLIST\n")
}

// Return final string
func (b *M3U8Builder) String() string {
	return b.builder.String()
}

// Write play item duration
func (b *M3U8Builder) ExtINF(seg *models.Segment) {

	// Calculate actual duration (e.g., 60.000 seconds)
	durationSeconds := float64(seg.EndTime - seg.StartTime)
	// if durationSeconds <= 0 {
	// 	durationSeconds = 60.0 // Failsafe
	// }

	// CRITICAL FIX: Clamp abnormal durations caused by camera disconnects
	// If the wall-clock time suggests a 2-hour file, clamp it to 60s because 
	// the physical video file can never be longer than our rotate interval.
	if durationSeconds <= 0 || durationSeconds > 120.0 {
		durationSeconds = 60.0 
	}

	b.builder.WriteString(fmt.Sprintf("#EXTINF:%.3f,%d\n", durationSeconds, seg.StartTime))
}

func (b *M3U8Builder) CalculateMaxDuration(segs []*models.Segment) int64 {
	min := int64(math.MaxInt)
	max := int64(0)
	for _, seg := range segs {
		if (seg.StartTime < min) { min = seg.StartTime }
		if (seg.EndTime > max) { max = seg.EndTime }
	}
	return max - min;
}

func (b *M3U8Builder) XSetTargetDurationFor(segs []*models.Segment) {
	b.XTargetDuration(int(b.CalculateMaxDuration(segs)))
}

func (b *M3U8Builder) FeedSegment(seg *models.Segment) {

	b.NL()
	b.ExtINF(seg)

	// 	// The actual URL VLC will call to get the video bytes.
	// 	// It points directly to our previously built Playback API!
	apiURI := utils.PathForCameraPlayMSURL(b.camID, seg.StartTime)
	segmentURL := fmt.Sprintf("%s%s\n", b.baseURL, apiURI)
	b.builder.WriteString(segmentURL)

}

func (b *M3U8Builder) FeedVODSegment(seg *models.Segment) {

	b.NL()
	b.XDiscontinuity()
	b.ExtINF(seg)

	// 	// The actual URL VLC will call to get the video bytes.
	// 	// It points directly to our previously built Playback API!
	apiURI := utils.PathForCameraTSPlayMSURL(b.camID, seg.StartTime)
	segmentURL := fmt.Sprintf("%s%s\n", b.baseURL, apiURI)
	b.builder.WriteString(segmentURL)

}