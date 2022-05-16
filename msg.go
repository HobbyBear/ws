package ws

import (
	"encoding/json"
)

type DataMsg struct {
	MsgId   string `json:"msg_id"`
	Content []byte `json:"content"`
}

//var (
//	dataMsgPool = sync.Pool{
//		New: func() interface{} {
//			return &DataMsg{}
//		},
//	}
//)

//func NewDataMsg() *DataMsg {
//	return dataMsgPool.Get().(*DataMsg)
//}

func (d DataMsg) MarshalJSON() []byte {
	data, _ := json.Marshal(d)
	return data
}
