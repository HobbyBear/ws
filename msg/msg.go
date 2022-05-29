package msg

import (
	"encoding/json"
	"github.com/gobwas/ws"
)

type PushType int

const (
	PushSingle PushType = iota
	PushGroup
	PushAll
)

type PushMsg struct {
	Type   PushType  `json:"type"`
	Uids   []string  `json:"uids,omitempty"`
	RoomId string    `json:"rid,omitempty"`
	Topic  string    `json:"topic,omitempty"`
	Data   []byte    `json:"data,omitempty"`
	WsType ws.OpCode `json:"ws_type"`
}

func (p *PushMsg) Marshal() []byte {
	data, _ := json.Marshal(p)
	return data
}

type ReqMsg struct {
	Path   string `json:"path"`
	Data   []byte `json:"data,omitempty"`
	Uid    string `json:"uid,omitempty"`
	RoomId string `json:"rid,omitempty"`
	Topic  string `json:"topic,omitempty"`
}

func (r *ReqMsg) Marshal() []byte {
	data, _ := json.Marshal(r)
	return data
}
