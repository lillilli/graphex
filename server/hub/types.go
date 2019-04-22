package hub

import "encoding/json"

// IncomingMessage - incoming messages format
type IncomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// OutgoingMessage - outgoing messages format
type OutgoingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// OutgoingResultMessage - outgoing result messages format
type OutgoingResultMessage struct {
	Type     string      `json:"type"`
	Success  bool        `json:"success"`
	Request  interface{} `json:"request_data"`
	ErrorMsg string      `json:"error,omitempty"`
}
