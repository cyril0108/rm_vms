package process

import (
	"nvr_core/db/models"
	"nvr_core/utils"
)

// StreamProfile represents the state of a specific stream on a specific worker
type StreamProfile struct {
    Source    string `json:"source"`
    LiveURL   string `json:"live_url"`
    WorkerID  int    `json:"worker_id"` // Moved here! Can be different for main/sub
    ChannelID int    `json:"channel_id"`
    VCodec    uint32 `json:"vcodec"`
    ACodec    uint32 `json:"acodec"`
    Status    string `json:"status"`    // e.g., "offline", "streaming", "failed"
}

// Camera represents the physical device for the API and Manager
type Camera struct {
    ID         int            `json:"id"`
    Active     bool           `json:"active"`
    MainStream *StreamProfile `json:"main_stream"`
    SubStream  *StreamProfile `json:"sub_stream"`
}

func (cam *Camera) NoSubStream() bool {
    return cam.SubStream == nil
}

func (cam *Camera) GetProfile(profile string) *StreamProfile {
    var pro *StreamProfile;

    switch profile {
    case "main":
        pro = cam.MainStream

    case "sub":
        pro = cam.SubStream

    default:
        return nil
    }
    return pro
}

// func NewCamera(camID int, url string, subUrl string) *Camera {
//     cam := &Camera{
//         ID: camID,
//         MainStream: *NewStreamProfile(url),
//         SubStream: *NewStreamProfile(subUrl),
//     }
//     return cam
// }

func NewCameraRuntime(c *models.Camera, masterKey []byte) *Camera {

    live := utils.URLForCameraLiveTSStream(c.ID, utils.SegmentMainProfile)
    cam := &Camera{
        ID: int(c.ID),
        Active: c.IsActive,
        MainStream: NewStreamProfile(c.AuthMainUrl(masterKey), live),
    }

    if c.SubStreamURL != "" {
        live := utils.URLForCameraLiveTSStream(c.ID, utils.SegmentSubProfile)
        cam.SubStream = NewStreamProfile(c.AuthSubUrl(masterKey), live)
    }

    return cam
}


func NewStreamProfile(url string, live string) *StreamProfile {
    pf := &StreamProfile{
        Source: url,
        LiveURL: live,
        WorkerID: -1,
        Status: "",
        ChannelID: -1,
    }
    return pf
}
