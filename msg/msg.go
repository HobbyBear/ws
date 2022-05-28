package msg

import "github.com/gobwas/ws"

type PushType int

const (
	PushSingle PushType = iota
	PushGroup
	PushAll
)

type PushMsg struct {
	Type    PushType  `json:"type"`
	Uid     string    `json:"uid,omitempty"`
	GroupId string    `json:"gid,omitempty"`
	Topic   string    `json:"topic,omitempty"`
	Data    []byte    `json:"data,omitempty"`
	WsType  ws.OpCode `json:"ws_type"`
}

type ReqMsg struct {
	Path    string `json:"path"`
	Data    []byte `json:"data,omitempty"`
	Uid     string `json:"uid,omitempty"`
	GroupId string `json:"groupId,omitempty"`
	Topic   string `json:"topic,omitempty"`
}
