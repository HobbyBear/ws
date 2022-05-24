package ws

import (
	"encoding/json"
)

type DataMsg struct {
	MsgType string `json:"msg_id,omitempty"`
	Content []byte `json:"content,omitempty"`
	Uid     string `json:"uid,omitempty"`
	GroupId string `json:"groupId,omitempty"`
	Topic   string `json:"topic,omitempty"`
}

func (d DataMsg) MarshalJSON() []byte {
	data, _ := json.Marshal(d)
	return data
}

// 一些预置的消息类型
const (
	SubMsgType   = "sub_topic" // 订阅的消息类型
	Login        = "login"
	UnSubMsgType = "un_sub_topic" // 取消订阅消息
)

func preDataHandler(conn *Conn, msg *DataMsg) {
	switch msg.MsgType {
	case SubMsgType:
		conn.topic = msg.Topic
	case Login:
		conn.Uid = msg.Uid
		conn.GroupId = msg.GroupId
	case UnSubMsgType:
		conn.topic = ""

	}
}

func postDataHandler(conn *Conn, msg *DataMsg) {
	//if msg.Ack == 1 {
	//	conn.WriteMsg(&RawMsg{WsMsgType: websocket.TextMessage, Content: DataMsg{MsgType: AckOk, AckId: msg.AckId}.MarshalJSON()})
	//}
}
