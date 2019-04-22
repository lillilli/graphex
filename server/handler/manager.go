package handler

import (
	"github.com/lillilli/graphex/server/handler/subscribe"
	"github.com/lillilli/graphex/server/hub"
	"github.com/lillilli/graphex/watcher"
)

type Manager interface {
	GetHander(msgType string) Handler
	HandleClientEvents(client *hub.Client)
}

type manager struct {
	handlers map[string]Handler
	emitter  hub.EventEmitter
	watcher  watcher.Watcher
}

func NewManager(emitter hub.EventEmitter, watcher watcher.Watcher) Manager {
	m := &manager{
		handlers: make(map[string]Handler),
		emitter:  emitter,
		watcher:  watcher,
	}

	m.initializeHandlers()
	return m
}

func (m *manager) initializeHandlers() {
	m.handlers["root_subscribe"] = &subscribe.RootSubscribeHandler{Emitter: m.emitter, Watcher: m.watcher}
	m.handlers["file_subscribe"] = &subscribe.FileSubscribeHandler{Emitter: m.emitter, Watcher: m.watcher}
}

// GetHander - returns handler by req type, if handler not exists it will return default handler
func (m *manager) GetHander(msgType string) Handler {
	handler, ok := m.handlers[msgType]

	if !ok {
		return &DefaultHandler{}
	}

	return handler
}

// HandleClientEvents - handle client events
func (m *manager) HandleClientEvents(client *hub.Client) {
	eventChannel := client.EventChannel()

	for event := range eventChannel {
		m.GetHander(event.Type).Handle(client, event.Data)
	}
}
