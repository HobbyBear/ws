package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type RouterHandlerReq struct {
	Conn    *Conn
	MsgId   string
	Content []byte
	WsMsgId int
}

type RouterHandler func(req *RouterHandlerReq)

type RouterMgr interface {
	GetRouterByMsgId(msgId string) RouterHandler
	RegHandler(msgId string, handler RouterHandler)
}

type defaultRouterMgr struct {
	routerMap map[string]RouterHandler
}

func (d *defaultRouterMgr) GetRouterByMsgId(msgId string) RouterHandler {
	return d.routerMap[msgId]
}

func (d *defaultRouterMgr) RegHandler(msgId string, handler RouterHandler) {
	d.routerMap[msgId] = handler
}

var (
	pongHandler = func(conn *Conn) {
		conn.wsConn.WriteControl(websocket.PongMessage, nil, time.Now().Add(time.Second))
	}
	pingHandler = func(conn *Conn) {
		conn.wsConn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second))
	}

	routerMgr RouterMgr = &defaultRouterMgr{
		routerMap: map[string]RouterHandler{},
	}

	requestHandler = func(req *Request) {
		switch req.MsgType {
		case websocket.PingMessage:
			pongHandler(req.Conn)
		case websocket.CloseMessage:
			req.Conn.Close()
			return
		case websocket.TextMessage, websocket.BinaryMessage:
			dataHandler(req.Conn, req.Data, req.MsgType)
		}
	}

	dataHandler = func(conn *Conn, data []byte, wsMsgType int) {
		type dataMsg struct {
			MsgId   string `json:"msg_id"`
			Content []byte `json:"content"`
		}
		var msg dataMsg
		err := json.Unmarshal(data, &msg)
		if err != nil {
			Errorf("data msg is invalid err=%s data=%s", err, string(data))
			return
		}
		if handler := routerMgr.GetRouterByMsgId(msg.MsgId); handler != nil {
			handler(&RouterHandlerReq{
				Conn:    conn,
				MsgId:   msg.MsgId,
				Content: msg.Content,
				WsMsgId: wsMsgType,
			})
		}
	}
	authHandler = func(data []byte) (*AuthMsg, bool) {
		var authMsg AuthMsg
		err := json.Unmarshal(data, &authMsg)
		if err != nil {
			return nil, false
		}
		return &authMsg, true
	}

	callOnConnStateChange = func(c *Conn, state ConnState) {
		Infof("连接状态: %s", state.String())
	}

	defaultConnMgr = &DefaultConnMgr{
		UidConnMap:   make(map[string][]*Conn),
		GroupConnMap: make(map[string][]*Conn),
		All:          make([]*Conn, 0),
		mux:          sync.Mutex{},
	}

	connMgr ConnMgr = defaultConnMgr
)

type AuthMsg struct {
	Uid     string
	GroupId string
}

type ConnMgr interface {
	Del(c *Conn)
	Add(c *Conn)
	GetConnByUid(uid string) []*Conn
	GetConnByGroupId(groupId string) []*Conn
	GetAllConn() []*Conn
}

type DefaultConnMgr struct {
	UidConnMap   map[string][]*Conn
	GroupConnMap map[string][]*Conn
	All          []*Conn
	mux          sync.Mutex
}

func (cm *DefaultConnMgr) GetAllConn() []*Conn {
	//TODO implement me
	panic("implement me")
}

func (cm *DefaultConnMgr) Add(c *Conn) {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	if len(c.Uid) != 0 {
		cm.UidConnMap[c.Uid] = append(cm.UidConnMap[c.Uid], c)
	}
	if len(c.GroupId) != 0 {
		cm.GroupConnMap[c.GroupId] = append(cm.UidConnMap[c.GroupId], c)
	}
	cm.All = append(cm.All, c)
}

func (cm *DefaultConnMgr) GetConnByUid(uid string) []*Conn {
	//TODO implement me
	panic("implement me")
}

func (cm *DefaultConnMgr) GetConnByGroupId(groupId string) []*Conn {
	//TODO implement me
	panic("implement me")
}

func (cm *DefaultConnMgr) Del(conn *Conn) {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	delete(cm.UidConnMap, conn.Uid)
	delete(cm.GroupConnMap, conn.GroupId)
}

func SetPongHandler(f func(conn *Conn)) {
	pongHandler = f
}

func SetPingHandler(f func(conn *Conn)) {
	pingHandler = f
}

func SetDataHandler(f func(conn *Conn, data []byte, wsMsgType int)) {
	dataHandler = f
}

func SetRouterMgr(r RouterMgr) {
	routerMgr = r
}

func GetRouterMgr() RouterMgr {
	return routerMgr
}

func SetRequestHandler(f func(req *Request)) {
	requestHandler = f
}

func SetAuthHandler(f func(data []byte) (*AuthMsg, bool)) {
	authHandler = f
}

func SetCallOnConnStateChange(f func(c *Conn, state ConnState)) {
	callOnConnStateChange = f
}

func SetDefaultConnMgr(mgr ConnMgr) {
	connMgr = mgr
}
