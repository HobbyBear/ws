package ws

import "github.com/gorilla/websocket"

func SendMsgToAll(data []byte) {
	for _, conn := range defaultConnMgr.GetAllConn() {
		conn.WriteMsg(&RawMsg{WsMsgType: websocket.TextMessage, Content: data})
	}
}

func SendMsg(data []byte, uid, groupId, topic string) {
	var (
		connList []*Conn
	)
	if len(uid) != 0 {
		connList = defaultConnMgr.GetConnByUid(uid)
	}
	if len(uid) == 0 && len(groupId) != 0 {
		connList = defaultConnMgr.GetConnByGroupId(groupId)
	}
	if len(uid) == 0 && len(groupId) == 0 && len(topic) != 0 {
		connList = defaultConnMgr.GetAllConn()
	}
	for _, conn := range connList {
		c := conn
		if len(uid) != 0 && conn.Uid != uid {
			continue
		}
		if len(groupId) != 0 && conn.GroupId != groupId {
			continue
		}
		if len(topic) != 0 && conn.topic != topic {
			continue
		}
		c.WriteMsg(&RawMsg{WsMsgType: websocket.TextMessage, Content: data})
	}
}
