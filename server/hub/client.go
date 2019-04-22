package hub

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lillilli/logger"
)

const (
	// Time allowed to read a message from the peer.
	readWait = 60 * time.Second

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Common error messages event
	errorMessageType = "error"

	// Error that throws, if we send message to a closed connection
	errorUseOfClosedConnection = "use of closed network connection"

	// Error that throws, when connection reset by peer
	errorConnectionResetByPeer = "connection reset by peer"

	// Error that throws, when pipe is broken
	errorBrokenPipe = "broken pipe"
)

// Client - middleman between the websocket connection and the hub.
type Client struct {
	hub  *Hub
	conn *websocket.Conn

	events       chan *IncomingMessage
	writeChannel chan interface{}

	ctx    context.Context
	cancel context.CancelFunc

	CurrentFile string

	disconnected bool
	log          logger.Logger
	sync.RWMutex
}

// NewClient - return new ws client instance
func NewClient(hub *Hub, conn *websocket.Conn, log logger.Logger) *Client {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &Client{
		hub:          hub,
		conn:         conn,
		events:       make(chan *IncomingMessage),
		writeChannel: make(chan interface{}),

		ctx:    ctx,
		cancel: cancel,

		log: log,
	}
}

// EventChannel - return client event channel
func (c *Client) EventChannel() chan *IncomingMessage {
	return c.events
}

// InitializeReadPump - initialize read client pump
func (c *Client) InitializeReadPump() {
	for {
		select {
		case <-c.ctx.Done():
			return

		default:
			msg := new(IncomingMessage)
			if err := c.conn.SetReadDeadline(time.Now().Add(readWait)); err != nil {
				c.log.Warnf("Setting read message deadline failed: %v", err)
			}

			code, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) || closeConnectionError(err) {
					c.hub.disconects <- c
					return
				}

				c.log.Warnf("Reading client message failed (code = %d): %v", code, err)
				c.hub.disconects <- c
				return
			}

			if err := json.Unmarshal(data, &msg); err != nil {
				c.sendBadEventDataFormat(data)
				continue
			}

			c.events <- msg
		}
	}
}

// InitializeWritePump - initialize write client pump
func (c *Client) InitializeWritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return

		case msg := <-c.writeChannel:
			if c.Disconnected() {
				return
			}

			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.log.Warnf("Setting write deadline failed: %v", err)
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				if closeConnectionError(err) {
					c.hub.disconects <- c
					return
				}

				c.log.Warnf("Sending json failed: %v", err)
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.log.Warnf("Setting write deadline for ping failed: %v", err)
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if !closeConnectionError(err) {
					c.log.Warnf("Sending ping message failed: %v", err)
				}

				c.hub.disconects <- c
				return
			}
		}
	}
}

func (c *Client) sendBadEventDataFormat(req []byte) {
	if c.Disconnected() {
		return
	}

	c.RLock()
	defer c.RUnlock()

	errMsg := "bad event data format"
	c.writeChannel <- &OutgoingResultMessage{Type: errorMessageType, Success: false, Request: req, ErrorMsg: errMsg}
}

// SendJSON - send json msg to client
func (c *Client) SendJSON(msgType string, v interface{}) {
	if c.Disconnected() {
		return
	}

	c.RLock()
	c.writeChannel <- &OutgoingMessage{msgType, v}
	c.RUnlock()

	time.Sleep(1 * time.Millisecond)
}

func (c *Client) setPingHandler() {
	c.conn.SetPingHandler(func(string) error {
		return c.conn.WriteMessage(websocket.PongMessage, nil)
	})
}

func (c *Client) setPongHandler() {
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
}

func closeConnectionError(err error) bool {
	return err == websocket.ErrCloseSent || strings.HasSuffix(err.Error(), errorConnectionResetByPeer) ||
		strings.HasSuffix(err.Error(), errorBrokenPipe) || strings.HasSuffix(err.Error(), errorUseOfClosedConnection)
}

// Close - close client connection and flush memory
func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()

	if c.disconnected {
		return
	}

	c.disconnected = true

	c.cancel()
	c.conn.Close()

	close(c.events)
	close(c.writeChannel)

	c.conn = nil
	c.hub = nil
	c = nil
}

// Disconnected - return disconnected client status
func (c *Client) Disconnected() bool {
	c.RLock()
	defer c.RUnlock()

	return c.disconnected
}
