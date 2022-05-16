package ws

import (
	"container/list"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
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
	sendPong = func(conn *Conn, data string) {
		conn.WriteMsg(&RawMsg{WsMsgType: websocket.PongMessage, Data: nil, DeadLine: time.Now().Add(time.Second)})
	}
	sendPing = func(conn *Conn) {
		conn.WriteMsg(&RawMsg{WsMsgType: websocket.PingMessage, Data: nil, DeadLine: time.Now().Add(time.Second)})
	}

	routerMgr RouterMgr = &defaultRouterMgr{
		routerMap: map[string]RouterHandler{},
	}

	dataHandler = func(conn *Conn, data []byte, wsMsgType int) {
		var msg = &DataMsg{}
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		err := json.Unmarshal(data, msg)
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
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
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
		All:          list.New(),
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
	All          *list.List
	mux          sync.Mutex
}

func (cm *DefaultConnMgr) GetAllConn() []*Conn {
	connList := make([]*Conn, 0)
	for i := cm.All.Front(); i != nil; i = i.Next() {
		connList = append(connList, i.Value.(*Conn))
	}
	return connList
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
	e := cm.All.PushFront(c)
	c.element = e
}

func (cm *DefaultConnMgr) GetConnByUid(uid string) []*Conn {
	return cm.UidConnMap[uid]
}

func (cm *DefaultConnMgr) GetConnByGroupId(groupId string) []*Conn {
	return cm.GroupConnMap[groupId]
}

func (cm *DefaultConnMgr) Del(conn *Conn) {
	cm.mux.Lock()
	defer cm.mux.Unlock()
	delete(cm.UidConnMap, conn.Uid)
	delete(cm.GroupConnMap, conn.GroupId)
	cm.All.Remove(conn.element)
}

func SetSendPongFunc(f func(conn *Conn, data string)) {
	sendPong = f
}

func SetSendPingFunc(f func(conn *Conn)) {
	sendPing = f
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

func SetAuthHandler(f func(data []byte) (*AuthMsg, bool)) {
	authHandler = f
}

func SetCallOnConnStateChange(f func(c *Conn, state ConnState)) {
	callOnConnStateChange = f
}

func SetDefaultConnMgr(mgr ConnMgr) {
	connMgr = mgr
}
