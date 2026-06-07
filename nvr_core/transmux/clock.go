package transmux

// StreamClock manages the synchronization and monotonicity of Audio and Video timestamps.
// It isolates the Dual-Clock anchor logic to prevent A/V skew and buffer deadlocks.
type StreamClock struct {
	videoFirstPTS int64
	audioFirstPTS int64
	hasVideoFirst bool
	hasAudioFirst bool
	
	lastVideoDTS  int64
	lastAudioDTS  int64
	lastPCR       int64

	bufferTicks   int64 // Configurable breathing room for the player (default 90000 = 1 sec)
}

func NewStreamClock() *StreamClock {
	return &StreamClock{
		bufferTicks: 90000,
	}
}

// NormalizeVideo calculates safe, monotonic timestamps and a PCR base for the video stream.
func (c *StreamClock) NormalizeVideo(pts, dts int64) (normPTS, normDTS, pcrBase int64) {
	if !c.hasVideoFirst {
		c.videoFirstPTS = dts
		c.hasVideoFirst = true
		c.lastVideoDTS = 0
		c.lastPCR = 0
	}

	// Dynamic Baseline: Prevent negative underflow if a packet arrives from the past
	if dts < c.videoFirstPTS {
		c.videoFirstPTS = dts
	}

	elapsedTicks := dts - c.videoFirstPTS
	elapsedPTS := pts - c.videoFirstPTS

	// Enforce strict monotonicity for hardware decoders
	if elapsedTicks <= c.lastVideoDTS {
		elapsedTicks = c.lastVideoDTS + 1
	}
	c.lastVideoDTS = elapsedTicks

	// Master Clock (PCR) uses the monotonic elapsed ticks
	pcrBase = elapsedTicks
	if pcrBase <= c.lastPCR {
		pcrBase = c.lastPCR + 1
	}
	c.lastPCR = pcrBase

	// Apply breathing buffer
	normDTS = elapsedTicks + c.bufferTicks
	normPTS = elapsedPTS + c.bufferTicks

	// Failsafe for missing DTS
	if dts == 0 && pts != 0 {
		normDTS = normPTS
	}

	// PTS must never mathematically occur before DTS
	if normPTS < normDTS {
		normPTS = normDTS
	}

	return normPTS, normDTS, pcrBase
}

// NormalizeAudio calculates safe, monotonic timestamps for the audio stream.
func (c *StreamClock) NormalizeAudio(pts, dts int64) (normPTS, normDTS int64) {
	if !c.hasAudioFirst {
		c.audioFirstPTS = dts
		c.hasAudioFirst = true
		c.lastAudioDTS = 0
	}

	if dts < c.audioFirstPTS {
		c.audioFirstPTS = dts
	}

	elapsedTicks := dts - c.audioFirstPTS
	elapsedPTS := pts - c.audioFirstPTS

	if elapsedTicks <= c.lastAudioDTS {
		elapsedTicks = c.lastAudioDTS + 1
	}
	c.lastAudioDTS = elapsedTicks

	normDTS = elapsedTicks + c.bufferTicks
	normPTS = elapsedPTS + c.bufferTicks

	if dts == 0 && pts != 0 {
		normDTS = normPTS
	}

	if normPTS < normDTS {
		normPTS = normDTS
	}

	return normPTS, normDTS
}