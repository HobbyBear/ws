package ws

import "github.com/gorilla/websocket"

func SendMsgToConn(uid string, data []byte) {
	connList := defaultConnMgr.GetConnByUid(uid)
	for _, conn := range connList {
		conn.WriteMsg(&RawMsg{WsMsgType: websocket.TextMessage, Data: data})
	}
}

func SendMsgToRoom(groupId string, data []byte) {
	connList := defaultConnMgr.GetConnByGroupId(groupId)
	for _, conn := range connList {
		conn.WriteMsg(&RawMsg{WsMsgType: websocket.TextMessage, Data: data})
	}
}

func SendMsgToAll(data []byte) {
	for _, conn := range defaultConnMgr.GetAllConn() {
		conn.WriteMsg(&RawMsg{WsMsgType: websocket.TextMessage, Data: data})
	}
}
