package transmux

type StreamClock struct {
	isAnchored   bool
	anchorBase   int64

	lastVideoDTS int64
	lastAudioDTS int64
	lastPCR      int64

	bufferTicks  int64
}

func NewStreamClock() *StreamClock {
	return &StreamClock{
		bufferTicks: 90000, // 1 second VLC breathing room
	}
}

func (c *StreamClock) NormalizeVideo(pts, dts int64) (int64, int64, int64) {
	if !c.isAnchored {
		c.anchorBase = dts
		c.isAnchored = true
	} else if dts < c.anchorBase {
		c.anchorBase = dts // Catch underflow if another stream starts earlier
	}

	elapsedTicks := dts - c.anchorBase
	elapsedPTS := pts - c.anchorBase

	if elapsedTicks <= c.lastVideoDTS {
		elapsedTicks = c.lastVideoDTS + 1
	}
	c.lastVideoDTS = elapsedTicks

	pcrBase := elapsedTicks
	if pcrBase <= c.lastPCR {
		pcrBase = c.lastPCR + 1
	}
	c.lastPCR = pcrBase

	normDTS := elapsedTicks + c.bufferTicks
	normPTS := elapsedPTS + c.bufferTicks

	if dts == 0 && pts != 0 { normDTS = normPTS }
	if normPTS < normDTS { normPTS = normDTS }

	return normPTS, normDTS, pcrBase
}

func (c *StreamClock) NormalizeAudio(pts, dts int64) (int64, int64) {
	if !c.isAnchored {
		c.anchorBase = dts
		c.isAnchored = true
	} else if dts < c.anchorBase {
		c.anchorBase = dts
	}

	elapsedTicks := dts - c.anchorBase
	elapsedPTS := pts - c.anchorBase

	if elapsedTicks <= c.lastAudioDTS {
		elapsedTicks = c.lastAudioDTS + 1
	}
	c.lastAudioDTS = elapsedTicks

	normDTS := elapsedTicks + c.bufferTicks
	normPTS := elapsedPTS + c.bufferTicks

	if dts == 0 && pts != 0 { normDTS = normPTS }
	if normPTS < normDTS { normPTS = normDTS }

	return normPTS, normDTS
}