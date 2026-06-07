package stream

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 64, // 64KB for large video keyframes
	CheckOrigin: func(r *http.Request) bool {
		return true // Lock this down in production
	},
}

type Client struct {
	hub                *Hub
	conn               *websocket.Conn
	subscriber         *Subscriber
}

// ServeWs handles websocket requests from the Vue frontend.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}

	client := &Client{
		hub:                hub,
		conn:               conn,
		subscriber: &Subscriber{
			Send:               make(chan StreamPacket, 256),
			WaitingForKeyframe: true, // Crucial for clean jmuxer initialization
		},
	}

	hub.Register <- client.subscriber

	// Start the write pump for binary frames
	go client.writePump()
	
	// Start the read pump to handle client disconnects gracefully
	go client.readPump()
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case packet, ok := <-c.subscriber.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(payloadForWebSocket(packet))

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c.subscriber
		c.conn.Close()
	}()
	for {
		// SPSC stream: We only read to detect close events from the browser
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}