package process

import "nvr_core/db/models"

// StreamProfile represents the state of a specific stream on a specific worker
type StreamProfile struct {
    URL       string `json:"url"`
    WorkerID  int    `json:"worker_id"` // Moved here! Can be different for main/sub
    ChannelID int    `json:"channel_id"`
    Status    string `json:"status"`    // e.g., "offline", "streaming", "failed"
}

// Camera represents the physical device for the API and Manager
type Camera struct {
    ID         int           `json:"id"`
    Active     bool          `json:"active"`
    MainStream *StreamProfile `json:"main_stream"`
    SubStream  *StreamProfile `json:"sub_stream"`
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

func NewCameraRuntime(c *models.Camera) *Camera {

    cam := &Camera{
        ID: int(c.ID),
        Active: c.IsActive,
        MainStream: NewStreamProfile(c.StreamURL),
    }

    if c.SubStreamURL != "" {
        cam.SubStream = NewStreamProfile(c.SubStreamURL)
    }

    return cam
}


func NewStreamProfile(url string) *StreamProfile {
    pf := &StreamProfile{
        URL: url,
        WorkerID: -1,
        Status: "",
        ChannelID: -1,
    }
    return pf
}
