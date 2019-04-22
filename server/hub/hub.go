package hub

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lillilli/logger"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	ctx context.Context

	connects   chan *Client
	disconects chan *Client

	emitter EventEmitter

	clients map[string]bool

	clientLogger logger.Logger
	log          logger.Logger
	sync.Mutex
}

// New - return new connection hub instance
func New(ctx context.Context, emitter EventEmitter) *Hub {
	hub := &Hub{
		ctx:     ctx,
		emitter: emitter,

		clients:    make(map[string]bool),
		connects:   make(chan *Client),
		disconects: make(chan *Client),

		clientLogger: logger.NewLogger("ws client"),
		log:          logger.NewLogger("ws hub"),
	}

	return hub
}

// NewClient - creates new client in ws hub
func (h *Hub) NewClient(conn *websocket.Conn) *Client {
	client := NewClient(h, conn, h.clientLogger)

	client.setPingHandler()
	client.setPongHandler()

	go client.InitializeReadPump()
	go client.InitializeWritePump()

	h.connects <- client

	return client
}

// Start - start working: handle connects, disconnects and send events
func (h *Hub) Start() {
	go h.run()
	go h.emitter.Start(h.ctx)
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.connects:
			h.Lock()
			h.clients[client.conn.RemoteAddr().String()] = true
			h.log.Infof("Client connect: %#v (client %d)", client.conn.RemoteAddr().String(), len(h.clients))
			h.emitter.AddSubscriberForRoot(client)
			h.Unlock()

		case client := <-h.disconects:
			if client == nil || client.conn == nil {
				return
			}

			h.Lock()

			if _, ok := h.clients[client.conn.RemoteAddr().String()]; ok {
				delete(h.clients, client.conn.RemoteAddr().String())
				h.log.Infof("Client disconnect: %#v (clients: %d)", client.conn.RemoteAddr().String(), len(h.clients))

				h.emitter.RemoveSubscriberForRoot(client)
				h.emitter.RemoveSubscriberForFile(client.CurrentFile, client)
				client.Close()
			}

			h.Unlock()
		}
	}
}
