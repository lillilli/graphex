package handler

import (
	"github.com/lillilli/graphex/server/hub"
)

// Handler - represents ws handler interface
type Handler interface {
	Handle(client *hub.Client, data []byte)
}

// DefaultHandler - default req handler
type DefaultHandler struct{}

// Handle - returns "unknown message type" error
func (h DefaultHandler) Handle(client *hub.Client, data []byte) {
	client.SendJSON("error", "unknown message type")
}
