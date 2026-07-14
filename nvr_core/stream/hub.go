package stream

import (
	"sync"
)

// Subscriber can be a WebSocket subscriber OR an HTTP response writer
type Subscriber struct {
	Send               chan StreamPacket
	WaitingForKeyframe bool
}

// StreamPacket encapsulates the raw video payload and its metadata.
type StreamPacket struct {
	MediaType  uint8  // 0 = Video, 1 = Audio
	CodecID    uint32
	IsKeyFrame bool
	PTS        int64
	DTS        int64
	Payload    []byte
}

type IFrameCache struct {
    sync.RWMutex
    lastIFrame *StreamPacket
    isSet      bool
}

// MediaType constants to match your C++ definitions
const (
	MediaTypeVideo uint8 = 0
	MediaTypeAudio uint8 = 1
)

// Hub maintains the set of active clients and broadcasts video frames.
type Hub struct {
	subscribers    map[*Subscriber]bool
	Broadcast      chan StreamPacket
	Register       chan *Subscriber
	Unregister     chan *Subscriber
	keyframeCache  *IFrameCache
	mu             sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan StreamPacket, 2048), // Buffered to handle high FPS bursts
		Register:   make(chan *Subscriber),
		Unregister: make(chan *Subscriber),
		subscribers:    make(map[*Subscriber]bool),
		keyframeCache: &IFrameCache{
			isSet: false,
		},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case subscriber := <-h.Register:
			h.subscribers[subscriber] = true
		case subscriber := <-h.Unregister:
			if _, ok := h.subscribers[subscriber]; ok {
				delete(h.subscribers, subscriber)
				close(subscriber.Send)
			}
		case packet := <-h.Broadcast:

			if packet.MediaType == MediaTypeVideo && packet.IsKeyFrame {
				h.cacheKeyframe(&packet)
			}

			for subscriber := range h.subscribers {
				// Late Joiner Logic: Drop frames until the first IDR Keyframe arrives
				if subscriber.WaitingForKeyframe && packet.MediaType == MediaTypeVideo {

					if !packet.IsKeyFrame && h.keyframeCache.isSet {

						select {
						case subscriber.Send <- *h.keyframeCache.lastIFrame:
						default:
							// Slow consumer detected: drop the subscriber to prevent blocking the Hub
							close(subscriber.Send)
							delete(h.subscribers, subscriber)
							continue // Skip to the next subscriber
						}

					}

					subscriber.WaitingForKeyframe = false

				}

				if packet.IsKeyFrame {
					h.cacheKeyframe(&packet)
				}

				select {
				case subscriber.Send <- packet:
				default:
					// Slow consumer detected: drop the subscriber to prevent blocking the Hub
					close(subscriber.Send)
					delete(h.subscribers, subscriber)
				}
			}
		}
	}
}

func (h *Hub) cacheKeyframe(packet *StreamPacket) {

	// Copies outer struct
	cachedFrame := packet

	// Allocate fresh memory and copy the bytes
	payloadCopy := make([]byte, len(packet.Payload))
	copy(payloadCopy, packet.Payload)
	cachedFrame.Payload = payloadCopy

	cache := h.keyframeCache
	cache.Lock()
	cache.lastIFrame = cachedFrame
	cache.isSet = true
	cache.Unlock()

}

func (h *Hub) GetKeyframeCache() *IFrameCache {
	return h.keyframeCache
}