package watcher

const CreateState = "CREATE"
const ModifyState = "MODIFY"
const RemoveState = "REMOVE"

type Event struct {
	Type   string    `json:"type"`
	Name   string    `json:"name"`
	Values *FileData `json:"data"`
}
