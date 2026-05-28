package process

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

    "nvr_core/db/models"
    "nvr_core/logger"
    "nvr_core/service"
    "nvr_core/shm"
    "nvr_core/stream"
    "nvr_core/utils"
)

const LOGSEP = "==============================================\n"

// Matches the JSON sent from C++
type WorkerResponse struct {
    Status    string `json:"status"`
    CamID     int    `json:"cam"`
    Profile   string `json:"profile"`
    ChannelID int    `json:"channel"`
    Size      int    `json:"size"`

    // the Segment Done event
    StartTime int64  `json:"start_time,omitempty"`
    EndTime   int64  `json:"end_time,omitempty"`
    FilePath  string `json:"file_path,omitempty"`
    SizeBytes int64  `json:"size_bytes,omitempty"`
}

// Worker represents a single C++ subprocess
type Worker struct {
    ID           int
    BinaryPath   string
    Cmd          *exec.Cmd
    Stdin        io.WriteCloser
    Stdout       io.ReadCloser
    Stderr       io.ReadCloser
    storagePath  string
    cameras      map[int]*Camera
    // SHM reader and stream hub
    shmReader    *shm.ReaderSHM
    streamHubs   map[int]*stream.Hub
    mu           sync.Mutex // Protects concurrent writes to Stdin
    dmu          utils.DebugMutex
    ingester     service.IngestService
    log          *logger.Logger
}

// NewWorker creates a struct but doesn't start the process yet
func NewWorker(id int, binaryPath string, ingester service.IngestService) *Worker {
    return &Worker{
        ID:         id,
        BinaryPath: binaryPath,
        cameras: make(map[int]*Camera),
        streamHubs: make(map[int]*stream.Hub),
        ingester:   ingester,
        log:        LOG.Lin("worker",id),
    }
}

func (w *Worker) handleStoppedStream(resp WorkerResponse) {

    w.mu.Lock()
    if cam, exists := w.cameras[resp.CamID]; exists {
        pf := cam.GetProfile(resp.Profile)
        if pf != nil {
            pf.Status = "Stopped"
            if pf.ChannelID > -1 {
                w.shmReader.StopChannel(resp.CamID, pf.ChannelID)
            }
        }
    }
    w.mu.Unlock()

    go func() {
        time.Sleep(8 * time.Second)
        fmt.Println(LOGSEP + "[Go][Worker] restarting cam:", resp.CamID, resp.Profile)
        w.RestartCam(resp.CamID, resp.Profile)
    }()

}

func (w *Worker) updateCameraStatus(resp WorkerResponse) {

    w.mu.Lock()
    defer w.mu.Unlock()

    cam, exists := w.cameras[resp.CamID]
    if !exists {
        return
    }

    pf := cam.GetProfile(resp.Profile)
    if pf != nil {
        pf.Status = resp.Status
    }

}

func (w *Worker) updateCameraSHMChannel(resp WorkerResponse) {

    fmt.Println(LOGSEP + "[Go][Worker] update SHM Channel cam:", resp.CamID, resp.ChannelID)

    w.updateCameraStatus(resp)

    w.mu.Lock()
    cam := w.cameras[resp.CamID]

    existingHub := w.streamHubs[resp.ChannelID]
    w.mu.Unlock()

    // Update SHM stream reader
    if(w.shmReader == nil) {
        fmt.Println("[Go][Worker] no shm reader for worker:", w.ID)
        return
    }

    pf := cam.GetProfile(resp.Profile)
    if pf != nil {

        w.mu.Lock()

        pf.ChannelID = resp.ChannelID
        hub := w.shmReader.StartChannel(resp.CamID, resp.ChannelID, existingHub)
        w.streamHubs[resp.ChannelID] = hub

        w.mu.Unlock()

    }


}

func (w *Worker) handleSegmentDone(resp WorkerResponse) {
    // --- DB HOOK ---
    seg := &models.Segment{
        CameraID:  int64(resp.CamID),
        Profile:   resp.Profile,
        StartTime: resp.StartTime,
        EndTime:   resp.EndTime,
        FilePath:  resp.FilePath,
        SizeBytes: resp.SizeBytes,
    }

    if seg.IsSubProfile() {
        if snap, err := utils.GenerateSnapshot(w.storagePath, seg); err != nil {
            seg.SnapshotPath = snap
        }
    }

    // Push it to the non-blocking Go channel
    // (Assuming you added `ingester *ingest.BatchIngester` to your Worker struct)
    if w.ingester != nil {
        w.ingester.Enqueue(seg)
    } else {
        fmt.Printf("\033[33m[Go][Worker %d] Warning: DB Ingester not configured, dropping segment metadata.\033[0m\n", w.ID)
    }
}

// ==================================

func (w *Worker) StreamHubForCam(camId int, profile string) *stream.Hub {
    cam := w.cameras[camId]

    pro := cam.GetProfile(profile)
    if pro == nil {
        fmt.Println("[Go][Worker][StreamHubForCam] incorrect profile ", profile)
        return nil
    }

    fmt.Println("[Go][Worker][StreamHubForCam] ", w.ID, camId, cam, pro.ChannelID)
    return w.streamHubs[pro.ChannelID]
}


// Start up SHM reader when we receive SHM data from cpp worker
func (w *Worker) startSHMReader(resp WorkerResponse) {

    fmt.Println("[Go][Worker] start shm reader", w.ID)

    // Launch the reader in the background so it doesn't block the command loop
    w.shmReader = shm.StartStreamReader(strconv.Itoa(w.ID), 10, resp.Size)

}

func (w *Worker) SetStoragePath(s string) {
    w.storagePath = s
}

// func (w *Worker) commandString() string {
//     return fmt.Sprintf("%s %s", w.BinaryPath, w.storagePath)
// }

// Start launches the C++ binary and sets up pipes
// This code will setup pipes and send WorkerID to
// cpp program, and should not be called twice.
func (w *Worker) Start(ctx context.Context) error {

    w.Cmd = exec.CommandContext(ctx, w.BinaryPath, w.storagePath)

    // Configure Graceful Termination (Crucial for /dev/shm cleanup)
    // By default, CommandContext sends a brutal SIGKILL when the context cancels.
    // We override this to send a SIGTERM first, giving your C++ destructors a chance to run.
    w.Cmd.Cancel = func() error {
        // log.Printf("[Worker %s] Sending SIGTERM for graceful shutdown...", workerID)
        return w.Cmd.Process.Signal(syscall.SIGTERM)
    }

    // Give the C++ worker 3 seconds to unmap memory and exit smoothly. 
    // If it hangs, Go will follow up with a SIGKILL to forcefully terminate it.
    w.Cmd.WaitDelay = 3 * time.Second

    // Setup Stdin Pipe
    if err := w.ConfigureCMDPipes(); err != nil {
        return err;
    }

    if err := w.Cmd.Start(); err != nil {
        return fmt.Errorf("worker %d start failed: %v", w.ID, err)
    }

    w.connectCMDIPC()
    w.SendWorkerID()

    // Launch the monitor in the background. 
    // This goroutine will sleep quietly until the C++ process dies.
    go w.monitorProcess(ctx)

    return nil
}

// SendCommand is a thread-safe way to write to the worker
func (w *Worker) SendCommand(cmd string) error {
    // w.mu.Lock()
    // defer w.mu.Unlock()

    if w.Stdin == nil {
        return fmt.Errorf("worker %d is not running", w.ID)
    }

    // Add newline because C++ uses std::getline
    _, err := io.WriteString(w.Stdin, cmd+"\n")
    return err
}

func (w *Worker) SendWorkerID() error {

    command := fmt.Sprintf("WORKER %d", w.ID)
    return w.SendCommand(command)

}


func (w *Worker) AssignCam(cam *Camera, profile string) error {

    w.mu.Lock()

    sp := cam.GetProfile(profile)
    if sp != nil {
        sp.WorkerID = w.ID
    }
    w.cameras[cam.ID] = cam

    w.mu.Unlock()

    return w.StartStreamProfile(cam.ID, profile, sp)

}

// func (w *Worker) StartCam(cam *Camera) error {

//     cam.Status = "starting"
//     command := fmt.Sprintf("START %d %s", cam.ID, cam.Main)
//     return w.SendCommand(command)

// }

func (w *Worker) StartStreamProfile(camID int, profile string, sp *StreamProfile) error {

    sp.Status = "starting"
    command := fmt.Sprintf("START %d %s %s", camID, profile, sp.URL)
    return w.SendCommand(command)

}

func (w *Worker) StopCamRecording(camID int, profile string) error {

    w.mu.Lock()
    defer w.mu.Unlock()

    cam, exists := w.cameras[camID]
    if !exists {
        return fmt.Errorf("camera %d not found", camID)
    }

    pro := cam.GetProfile(profile)
    if pro == nil {
        fmt.Println("[Go][Worker][StreamHubForCam] incorrect profile ", profile)
        return nil
    }

    pro.Status = "streaming"
    command := fmt.Sprintf("NORECORDING %d", cam.ID)


    return w.SendCommand(command)

}

func (w *Worker) RestartCam(camID int, profile string) error {

    w.mu.Lock()
    defer w.mu.Unlock()

    cam := w.cameras[camID]
    sp := cam.GetProfile(profile)
    if sp == nil {
        return fmt.Errorf("worker %d: no profile to restart", w.ID)
    }

    return w.StartStreamProfile(camID, profile, sp)

}


// Stop cleanly closes the pipe and waits
func (w *Worker) Stop() {
    if w.Stdin != nil {
        w.Stdin.Close() // Sends EOF to C++
    }
    if w.Stdout != nil {
        w.Stdout.Close() // Sends EOF to C++
    }
    if w.Stderr != nil {
        w.Stderr.Close() // Sends EOF to C++
    }
    if w.Cmd != nil {
        w.Cmd.Wait()
    }
}

/**
 * ======================================================
 * Getters
 * ======================================================
 */

func (w *Worker) GetCameras() []*Camera {
    w.mu.Lock()         // Lock before reading the map
    defer w.mu.Unlock() // Ensure it unlocks when the function finishes

    cameraList := make([]*Camera, 0, len(w.cameras))
    for _, camera := range w.cameras {
        cameraList = append(cameraList, camera)
    }
    return cameraList
}

func (w *Worker) GetSHMReader() *shm.ReaderSHM {
    w.mu.Lock()         // Lock before reading the map
    defer w.mu.Unlock() // Ensure it unlocks when the function finishes
    return w.shmReader
}

/**
 * ======================================================
 * Logging and IPC Response
 * ======================================================
 */

// handleCMDResponse parses the action from the C++ worker and routes it.
func (w *Worker) handleCMDResponse(resp WorkerResponse) {
    switch resp.Status {
    case "stopped":
        w.handleStoppedStream(resp)

    case "recording":
        fallthrough
    case "starting":
        w.updateCameraStatus(resp)

    case "streaming":
        w.updateCameraSHMChannel(resp)

    case "shm":
        w.startSHMReader(resp)

    case "segment_done":
        w.handleSegmentDone(resp)


    default:
        // Catch-all for unexpected IPC messages
        fmt.Printf("\033[33m[Go][Worker %d] Received unknown status from C++: '%s'\033[0m\n", w.ID, resp.Status)
    }
}

func (w *Worker) connectCMDIPC() {

    // Listen for STDOUT (JSON Responses)
    go func() {
        scanner := bufio.NewScanner(w.Stdout)
        for scanner.Scan() {
            text := scanner.Text()
            var resp WorkerResponse
            // If it's valid JSON, print prettily
            if err := json.Unmarshal([]byte(text), &resp); err == nil {
                w.handleCMDResponse(resp);
            } else {
                fmt.Printf("\033[38;2;0;200;0m%s\033[0m\n", text)
            }
        }
    }()

    // Listen for STDERR (Logs)
    go func() {
        scanner := bufio.NewScanner(w.Stderr)
        for scanner.Scan() {
            // Print C++ logs in Red color
            fmt.Printf("\033[31m%s\033[0m\n", scanner.Text())
        }
    }()

}

// Setup ALL Pipes
// This must be done BEFORE starting the command
func (w *Worker) ConfigureCMDPipes() error {
    stdin, err := w.Cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("worker %d stdin error: %v", w.ID, err)
    }
    w.Stdin = stdin

    stdout, err := w.Cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("worker %d stdout error: %v", w.ID, err)
    }
    w.Stdout = stdout

    stderr, err := w.Cmd.StderrPipe()
    if err != nil {
        return fmt.Errorf("worker %d stderr error: %v", w.ID, err)
    }
    w.Stderr = stderr

    return nil
}


/**
 * ======================================================
 * Monitor and Recover
 * ======================================================
 */
// monitorProcess waits for the C++ binary to exit and triggers a restart if it was unexpected.
func (w *Worker) monitorProcess(ctx context.Context) {

    // Wait blocks until the C++ process exits.
    err := w.Cmd.Wait()

    // Check if this was an intentional shutdown from the Go Manager
    if ctx.Err() != nil {
        fmt.Printf("[Go][Worker %d] Shut down gracefully.\n", w.ID)
        return
    }

    // If we reach here, the C++ process died unexpectedly!
    if err != nil {
        // Log in red for visibility
        fmt.Printf("\033[31m[Go][Worker %d] CRASH DETECTED: %v\033[0m\n", w.ID, err)
    } else {
        fmt.Printf("\033[31m[Go][Worker %d] Exited unexpectedly with code 0.\033[0m\n", w.ID)
    }

    w.recoverWorker(ctx)

}

// recoverWorker attempts to restart the C++ binary and restore its previous state
func (w *Worker) recoverWorker(ctx context.Context) {
    fmt.Printf("[Go][Worker %d] Attempting to restart in 3 seconds...\n", w.ID)

    // Add a small backoff delay. If the C++ worker is crashing instantly on startup
    // (e.g., due to a bad config), this prevents an infinite CPU-burning crash loop.
    time.Sleep(3 * time.Second)

    var shmPath string

    // CLEANUP: Destroy the old SHM reader and its goroutines
    if w.shmReader != nil {
        shmPath = w.shmReader.FilePath()
        w.shmReader.Close()
        w.shmReader = nil
    }

    // WIPE: Delete the corrupted shared memory file from the OS
    if shmPath != "" {
        if err := os.Remove(shmPath); err != nil {
            if !os.IsNotExist(err) {
                fmt.Printf("\033[33m[Go][Worker %d] Warning: failed to unlink SHM %s: %v\033[0m\n", w.ID, shmPath, err)
            }
        } else {
            fmt.Printf("[Go][Worker %d] Successfully wiped corrupted shared memory at %s\n", w.ID, shmPath)
        }
    } else {
        fmt.Printf("\033[33m[Go][Worker %d] Warning: No SHM path found to wipe.\033[0m\n", w.ID)
    }

    // Restart the process
    if err := w.Start(ctx); err != nil {
        fmt.Printf("[Go][Worker %d] Failed to restart: %v\n", w.ID, err)
        // In a production app, you might want to retry with an exponential backoff here
        return
    }

    //=============================================
    // This used to be the section to restart all
    // cameras. Since we do not know the profile
    // for those camera in this conext. It needs
    // another way to recover the cameras.
    //=============================================
    // Copy camera list, preventing infinite lock
    // camsToRestart := utils.CopyMapValues(w.cameras, &w.mu)

    // // Restore State: Re-assign all cameras that this worker was managing
    // w.mu.Lock()
    // defer w.mu.Unlock()
    // for _, cam := range camsToRestart {
    //     fmt.Printf("[Go][Worker %d] Restoring stream for Cam %d...\n", w.ID, cam.ID)
    //     w.StartCam(cam)
    // }
}