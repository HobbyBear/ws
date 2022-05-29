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

// ProduceType 与业务层沟通的消息类型
type ProduceType int

const (
	Push ProduceType = iota // 消息用于推送
	Api                     // 消息用于做功能性函数，比如踢人，禁言等。
)

type PushData struct {
	PType  PushType  `json:"ptype,omitempty"`
	Uids   []string  `json:"uids,omitempty"`
	RoomId string    `json:"rid,omitempty"`
	Topic  string    `json:"topic,omitempty"`
	Data   []byte    `json:"data,omitempty"`
	WsType ws.OpCode `json:"ws_type,omitempty"`
}

type ApiData struct {
	Param []byte `json:"param,omitempty"`
	Path  string `json:"path,omitempty"`
}

type ProduceMsg struct {
	ProduceType ProduceType `json:"type"`
	PushData
	ApiData
}

func (p *ProduceMsg) IsPush() bool {
	return p.ProduceType == Push
}

func (p *ProduceMsg) IsApi() bool {
	return p.ProduceType == Api
}

func (p *ProduceMsg) Marshal() []byte {
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
