package subscribe

import (
	"encoding/json"

	"github.com/lillilli/graphex/server/hub"
	"github.com/lillilli/graphex/watcher"
)

// FileSubscribeHandler - file subscribe handler
type FileSubscribeHandler struct {
	Watcher watcher.Watcher
	Emitter hub.EventEmitter
}

type FileSubscribeParams struct {
	FileName string `json:"name"`
}

func (h FileSubscribeHandler) Handle(client *hub.Client, data []byte) {
	params := new(FileSubscribeParams)

	if err := json.Unmarshal(data, &params); err != nil {
		client.SendJSON("file_subscribe", "parsing params failed")
		return
	}

	b, err := h.Watcher.FileState(params.FileName)
	if err != nil {
		client.SendJSON("file_subscribe", "reading file failed")
		return
	}

	h.Emitter.RemoveSubscriberForFile(client.CurrentFile, client)
	h.Emitter.AddSubscriberForFile(params.FileName, client)
	client.SendJSON("file_subscribe", b)
	client.CurrentFile = params.FileName
}
