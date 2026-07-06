package shm

import (
	"fmt"
	"log"
	"nvr_core/logger"
	"nvr_core/stream"

	"sync"
	"sync/atomic"
	"time"
)

type ReaderSHM struct {
	log            *logger.Logger
	workerName     string
	worker         *WorkerSHM
	camChannels    map[int]int
	channelHub     map[int]*stream.Hub
	channelStopper map[int]*atomic.Bool
	wg             sync.WaitGroup //

	debugCount      int
}

// StartStreamReader connects to a Worker's SHM and starts reading its channels.
// workerName: e.g., "worker1" (Matches the C++ name creation logic)
// numChannels: 16 (Based on your sharding strategy)
// bufferSize: 1500000 (The 1.5MB calculated earlier)
func StartStreamReader(workerName string, numChannels int, bufferSize int) (*ReaderSHM) {
	shmName := "shm" + workerName
	var shm *WorkerSHM
	var err error

	// Retry loop to wait for C++ to create the file
	for i := 0; i < 50; i++ {
		shm, err = ConnectSharedMemory(shmName, numChannels, bufferSize)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond) // Wait 100ms and try again
	}

	if err != nil {
		LOG.Error(fmt.Sprintf("[Go Manager] Failed to connect to SHM %s after retries: %v\n", shmName, err))
		return nil
	}

	LOG.Info(fmt.Sprintf("[Go Manager] Successfully connected to %s. Spawning the reader...\n", shmName))

	reader := ReaderSHM{
		log: LOG.Prefix("[worker"+workerName+"]"),
		workerName: workerName,
		worker: shm,
		camChannels: make(map[int]int),
		channelHub: make(map[int]*stream.Hub),
		channelStopper: make(map[int]*atomic.Bool),
	}

	// TODO: In production, you would attach a context.Context here 
	// to handle graceful shutdown and call shm.Close() when the worker dies.

	// context.Context

	LOG.Info(fmt.Sprintf("[Go Manager] %s reader spawned\n", shmName))

	return &reader;
}

func (r *ReaderSHM) FilePath() string {
	if r.worker != nil && r.worker.file != nil {
		return r.worker.file.Name() // Returns exactly what was passed to os.OpenFile
	}
	return ""
}

func (r *ReaderSHM) StartChannel(camID int, channelID int, existingHub *stream.Hub) *stream.Hub {

	rb := r.worker.Channels[channelID]
	if(rb == nil) {
		LOG.Error(fmt.Sprintf("[startChannel][%d] no ring buffer!\n", channelID))
		return nil;
	}

	r.camChannels[camID] = channelID
	hub := existingHub
	if(hub == nil) {

		LOG.Info(fmt.Sprintf("[startChannel][%d] streamer starting!\n", channelID))
		hub = stream.NewHub()
		go hub.Run()

	}

	r.channelHub[channelID] = hub
	stopper := &atomic.Bool{}
	stopper.Store(false)
	r.channelStopper[channelID] = stopper

	r.wg.Add(1)
	go r.readChannelLoop(stopper, channelID, rb, hub)

	return hub

}

func (r *ReaderSHM) StopChannel(camID int, channelID int) {
	stopper := r.channelStopper[channelID]
	if stopper == nil {
		return
	}
	stopper.Store(true);
	delete(r.camChannels, camID);
}

// readChannelLoop continuously polls a specific RingBuffer for new frames
func (r *ReaderSHM) readChannelLoop(stop *atomic.Bool, channelID int, rb *RingBuffer, bc *stream.Hub) {

	defer r.wg.Done() // Ensure counter decrements when loop exits safely

	// log.Printf("[Go][shm][readChannelLoop] Started reading loop for %s Channel %d\n", r.workerName, channelID)

	for {

		if(stop.Load()) {
			break;
		}

		// Attempt to read a frame
		// frameData, timestamp, ok := rb.ReadFrame()
		meta, frameData, ok := rb.ReadFrame()

		// fmt.Printf("Read %d bytes from SHM for Worker %s Channel %d\n", len(frameData), r.workerName, channelID)

		if !ok {
			// Buffer is empty.
			// Sleep for 1 millisecond to prevent the Goroutine from pegging the CPU to 100%.
			// At 30 FPS, a frame arrives roughly every ~33ms, so 1ms polling is highly responsive.
			time.Sleep(1 * time.Millisecond)
			continue
		}

		// --- ZERO-COPY READ SUCCESS ---
		// frameData now holds the raw video bytes (e.g., H.264 NAL units).
		// You can route this data to disk, a WebSocket, or a WebRTC pipeline.

		f := stream.StreamPacket {
			MediaType:  meta.MediaType,
			CodecID:    meta.CodecID,
			IsKeyFrame: meta.IsKeyFrame == 1,
			PTS:        meta.PTS,
			DTS:        meta.DTS,

			// Create a copy of the payload so the ring buffer can overwrite the old memory safely
			Payload:    frameData,
		}

		// if r.checkDebugCount() {
		// 	r.log.Info("[#"+strconv.Itoa(r.debugCount)+"] reading packet", "size", meta.FrameSize)
		// }

		bc.Broadcast <- f
		// fileDumpTest(frameData, r.workerName, channelID)

	}

}

// 
func (r *ReaderSHM) checkDebugCount() bool {
	r.debugCount++
	return r.debugCount < 10 || (r.debugCount % 10000) < 10
}

// GetWorkerMetrics polls the atomic state of all active channels for this worker
func (r *ReaderSHM) GetWorkerMetrics() *WorkerMetrics {
	metrics := &WorkerMetrics{
		WorkerID: r.workerName,
		Channels: make(map[int]*ChannelMetrics),
	}

	for camID, channelID := range r.camChannels {
		if rb := r.worker.Channels[channelID]; rb != nil {
			metrics.Channels[channelID] = rb.GetMetrics(camID, channelID)
		}
	}

	return metrics
}

// Close safely stops all polling goroutines and unmaps the shared memory
func (r *ReaderSHM) Close() {
	log.Printf("[shm.reader] Closing SHM Reader for %s\n", r.workerName)

	// Signal all readChannelLoop goroutines to exit
	for _, stopper := range r.channelStopper {
		stopper.Store(true)
	}

	// Wait for all goroutines to finish their current loop and exit
	r.wg.Wait()

	// Now we can unmap the memory for worker
	if r.worker != nil {
		r.worker.Close()
	}
}

// func fileDumpTest(frame []byte, worker string, channel int) {

// 	f, err := os.OpenFile(fmt.Sprintf("/app/recordings/debug_worker%v_cam_%d.264", worker, channel), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer f.Close()

// 	if _, err := f.Write(frame); err != nil {
// 		log.Println("Write error:", err)
// 	}

// }