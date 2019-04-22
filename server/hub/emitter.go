package hub

import (
	"context"
	"sync"

	"github.com/lillilli/logger"

	"github.com/lillilli/graphex/server/events"
	"github.com/lillilli/graphex/watcher"
)

// EventEmitter - hub event emitter interface
// send event updates to clients
type EventEmitter interface {
	Start(ctx context.Context)

	AddSubscriberForRoot(client *Client)
	AddSubscriberForFile(fileName string, client *Client)

	RemoveSubscriberForRoot(client *Client)
	RemoveSubscriberForFile(fileName string, client *Client)
}

type eventEmitter struct {
	watcher watcher.Watcher

	subscribersOnRoot []*Client
	subscribersOnFile map[string][]*Client

	log logger.Logger
	sync.Mutex
}

// NewEventEmitter - return new hub event emitter instance
func NewEventEmitter(watcher watcher.Watcher) EventEmitter {
	return &eventEmitter{
		watcher:           watcher,
		subscribersOnRoot: make([]*Client, 0),
		subscribersOnFile: make(map[string][]*Client),
		log:               logger.NewLogger("hub event emitter"),
	}
}

func (e *eventEmitter) Start(ctx context.Context) {
	e.log.Info("Starting ...")
	go e.startEmitFileUpdates(ctx)
}

func (e *eventEmitter) startEmitFileUpdates(ctx context.Context) {
	updatesChannel := e.watcher.UpdatesChannel()

	for {
		select {
		case <-ctx.Done():
			return
		case data := <-updatesChannel:
			if data.Type == watcher.CreateState || data.Type == watcher.RemoveState {
				go e.sendEventForRoot()
				continue
			}

			if data.Type == watcher.ModifyState {
				go e.sendEventForFile(data)
			}
		}
	}
}

func (e *eventEmitter) sendEventForRoot() {
	e.Lock()

	for _, client := range e.subscribersOnRoot {
		client.SendJSON(events.RootSubscribeEvent, e.watcher.State())
	}

	e.Unlock()
}

func (e *eventEmitter) sendEventForFile(data *watcher.Event) {
	e.Lock()
	defer e.Unlock()

	subscribers, ok := e.subscribersOnFile[data.Name]
	if !ok {
		return
	}

	for _, client := range subscribers {
		client.SendJSON(events.FileSubscribeEvent, data.Values)
	}
}

func (e *eventEmitter) AddSubscriberForRoot(client *Client) {
	e.Lock()
	e.subscribersOnRoot = append(e.subscribersOnRoot, client)
	e.Unlock()

	client.SendJSON("root_subscribe", e.watcher.State())
}

func (e *eventEmitter) AddSubscriberForFile(fileName string, client *Client) {
	e.Lock()

	if _, ok := e.subscribersOnFile[fileName]; !ok {
		e.subscribersOnFile[fileName] = make([]*Client, 0)
	}

	e.subscribersOnFile[fileName] = append(e.subscribersOnFile[fileName], client)
	e.Unlock()
}

func (e *eventEmitter) RemoveSubscriberForRoot(client *Client) {
	e.Lock()
	defer e.Unlock()

	for i, subscriber := range e.subscribersOnRoot {
		if client == subscriber {
			e.subscribersOnRoot = append(e.subscribersOnRoot[:i], e.subscribersOnRoot[i+1:]...)
		}
	}
}

func (e *eventEmitter) RemoveSubscriberForFile(fileName string, client *Client) {
	e.Lock()
	defer e.Unlock()

	subscribers, ok := e.subscribersOnFile[fileName]
	if !ok {
		return
	}

	for i, subscriber := range subscribers {
		if client == subscriber {
			subscribers = append(subscribers[:i], subscribers[i+1:]...)
		}
	}

	e.subscribersOnFile[fileName] = subscribers
}
