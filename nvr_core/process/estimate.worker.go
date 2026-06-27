package process

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// EstimationWorker inherits your base Worker but adds IPC routing for probes.
type EstimationWorker struct {
	*Worker // Inherits Stdin, Stdout, Start(), Stop(), etc.

	// pendingProbes maps a task key (e.g., "1_main") to a channel waiting for the result.
	pendingProbes sync.Map
}

// NewEstimationWorker initializes the dedicated worker.
func NewEstimationWorker(baseWorker *Worker) *EstimationWorker {
	ew := &EstimationWorker{
		Worker: baseWorker,
	}
	ew.hookCMDResponse()
	return ew
}

func (ew *EstimationWorker) hookCMDResponse() {
	go func() {
		scanner := bufio.NewScanner(ew.Worker.Stdout)
		for scanner.Scan() {
			text := scanner.Text()
			ew.HandleEstimationStdout(text)
		}
	}()
}

// RequestProbe sends a JSON command via Stdin and blocks until the specific C++ response returns.
func (ew *EstimationWorker) RequestProbe(ctx context.Context, camID int, profile string, rtspURL string) (float64, error) {

	// Create a unique composite key for this specific stream
	taskKey := fmt.Sprintf("%d_%s", camID, profile)

	// Create a buffered channel to receive the result
	resultChan := make(chan float64, 1)

	// Store the channel in our routing map
	ew.pendingProbes.Store(taskKey, resultChan)

	// Ensure we always clean up the map when this function exits
	defer ew.pendingProbes.Delete(taskKey)

	// Send to C++ via Stdin (Must lock the worker's Mutex to prevent interleaved Stdin writes)
	ew.mu.Lock()
	err := ew.SendCommand(fmt.Sprintf("PROBE %d %s %s", camID, profile, rtspURL))
	ew.mu.Unlock()

	if err != nil {
		return 0, fmt.Errorf("failed to send probe command: %w", err)
	}

	// Wait for the result from Stdout OR for the context to time out
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("probe request timed out for %s: %w", taskKey, ctx.Err())
	case estMB := <-resultChan:
		return estMB, nil
	}
}

// HandleEstimationStdout should be called inside this worker's stdout scanner loop
func (ew *EstimationWorker) HandleEstimationStdout(line string) {
	var resp WorkerResponse
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		LOG.Info("[EstimationWorker]" + line)
		return // Not a JSON line, ignore
	}

	if resp.Status == "ess" {
		// Reconstruct the routing key from the C++ response
		taskKey := fmt.Sprintf("%d_%s", resp.CamID, resp.Profile)

		// Look up the channel in the map
		if ch, exists := ew.pendingProbes.Load(taskKey); exists {
			// Route the C++ data back to the Go function that requested it
			ch.(chan float64) <- resp.EstimatedMB
		}
	}
}