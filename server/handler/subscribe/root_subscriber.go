package subscribe

import (
	"github.com/lillilli/graphex/server/hub"
	"github.com/lillilli/graphex/watcher"
)

// RootSubscribeHandler - root subscribe handler
type RootSubscribeHandler struct {
	Watcher watcher.Watcher
	Emitter hub.EventEmitter
}

func (h RootSubscribeHandler) Handle(client *hub.Client, data []byte) {
	h.Emitter.RemoveSubscriberForFile(client.CurrentFile, client)
	client.SendJSON("root_subscribe", h.Watcher.State)
}
