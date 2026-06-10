package utils

import (
	"fmt"
	"strconv"
)

// Should be identical with db/models/segment.go
const SegmentMainProfile = "main"
const SegmentSubProfile = "sub"
const SegmentSnapshotProfile = "snapshot"

// Return 1 for main, 0 for sub, -1 for unknown
func IsMainProfile(profile string) int {
	switch profile {
	case SegmentMainProfile:
		return 1;

	case SegmentSubProfile:
		fallthrough
	case SegmentSnapshotProfile:
		return 0;

	default:
		return -1
	}
}

func SanitizeProfile(profile string) string {
	switch profile {
	case SegmentMainProfile:
		fallthrough
	case SegmentSubProfile:
		return profile

	default:
		return SegmentMainProfile
	}
}

func CamID2Str(camID int64) string {
	return strconv.Itoa(int(camID))
}

func Str2CamID(str string) (int64, error) {
	if id, err := Str2Int(str); err == nil {
		return int64(id), err
	} else {
		return 0, err
	}
}

func Str2Int(str string) (int, error) {
	return strconv.Atoi(str)
}

// func PathForCameraPlayURL(camID int64, time int64) string {
// 	return fmt.Sprintf("/api/cameras/%d/play?time=%d", camID, time)
// }

// func PathForCameraTSPlayURL(camID int64, time int64) string {
// 	return fmt.Sprintf("/api/cameras/%d/play/ts?time=%d", camID, time)
// }

// /api/cameras/{id}/snapshot?time=1711000050
func PathForPlaybackSnapshotURL(camID int64, time int64) string {
	return fmt.Sprintf("/api/cameras/%d/snapshot?time=%d", camID, time)
}

// /api/cameras/{id}/snapshot?mstime=1711000050
func PathForPlaybackSnapshotMSURL(camID int64, time int64) string {
	return fmt.Sprintf("/api/cameras/%d/snapshot?mstime=%d", camID, time)
}

func PathForCameraPlayURL(camID int64, time int64) string {
	return fmt.Sprintf("/api/cameras/%d/play?time=%d", camID, time)
}

func PathForCameraPlayMSURL(camID int64, time int64) string {
	return fmt.Sprintf("/api/cameras/%d/play?mstime=%d", camID, time)
}

func PathForCameraTSPlayMSURL(camID int64, time int64) string {
	return fmt.Sprintf("/api/cameras/%d/play/ts?mstime=%d", camID, time)
}

// HandleGetPlaylist expects: GET /api/cameras/{cam_id}/playlist.m3u8?start=1711000000&end=1711003600
func PathForCameraPlaylistURL(camID int64, start int64, end int64) string {
	return fmt.Sprintf("/api/cameras/%d/playlist.m3u8?start=%d&end=%d", camID, start, end)
}

func PathForCameraVODPlaylistURL(camID int64, start int64, end int64) string {
	return fmt.Sprintf("/api/cameras/%d/playlist/ts.m3u8?start=%d&end=%d", camID, start, end)
}


// For MPEG-TS live stream
func URLForCameraLiveTSStream(baseUrl string, camID int64) string {
	return fmt.Sprintf("%s/live/camera/%d", baseUrl, camID)
}

// For WebSocket live stream
func URLForCameraWSStream(host string, camID int64) string {
	return fmt.Sprintf("ws://%s/ws/stream/%d", host, camID)
}