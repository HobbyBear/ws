package ws

import (
	"container/list"
	"github.com/gobwas/ws"
	jsoniter "github.com/json-iterator/go"
	"sync"
	"time"
)

type RouterHandlerReq struct {
	Conn    *Conn
	MsgId   string
	Content string
	WsMsgId ws.OpCode
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
		conn.WriteMsg(&RawMsg{WsMsgType: ws.OpPong, Content: nil, DeadLine: time.Now().Add(time.Second)})
	}
	sendPing = func(conn *Conn) {
		conn.WriteMsg(&RawMsg{WsMsgType: ws.OpPong, Content: nil, DeadLine: time.Now().Add(time.Second)})
	}

	routerMgr RouterMgr = &defaultRouterMgr{
		routerMap: map[string]RouterHandler{},
	}

	dataHandler = func(conn *Conn, data []byte, wsMsgType ws.OpCode) {
		conn.server.conTicker.AddTickConn(conn)
		if wsMsgType == ws.OpClose {
			conn.Close("对端主动关闭")
			return
		}
		if wsMsgType == ws.OpPing {
			sendPing(conn)
			return
		}

		if len(data) == 0 {
			return
		}

		var (
			msg  = &DataMsg{}
			json = jsoniter.ConfigCompatibleWithStandardLibrary
			err  error
		)
		err = json.Unmarshal(data, msg)
		if err != nil {
			Errorf("data msg is invalid err=%s data=%s", err, string(data))
			return
		}

		preDataHandler(conn, msg)
		if handler := routerMgr.GetRouterByMsgId(msg.MsgType); handler != nil {
			handler(&RouterHandlerReq{
				Conn:    conn,
				MsgId:   msg.MsgType,
				Content: msg.Content,
				WsMsgId: wsMsgType,
			})
		}
		postDataHandler(conn, msg)
	}

	callOnConnStateChange = func(c *Conn, state ConnState, reason string) {

	}

	defaultConnMgr = &DefaultConnMgr{
		UidConnMap:   make(map[string][]*Conn),
		GroupConnMap: make(map[string][]*Conn),
		All:          list.New(),
		mux:          sync.Mutex{},
	}

	connMgr ConnMgr = defaultConnMgr
)

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
	if len(c.uid) != 0 {
		cm.UidConnMap[c.uid] = append(cm.UidConnMap[c.uid], c)
	}
	if len(c.groupId) != 0 {
		cm.GroupConnMap[c.groupId] = append(cm.UidConnMap[c.groupId], c)
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
	if connList, ok := cm.UidConnMap[conn.uid]; ok {
		for i, c := range connList {
			if c.cid == conn.cid {
				index := len(connList) - 1
				if i+1 < len(connList)-1 {
					index = i + 1
				}
				connList = append(connList[:i], connList[index:]...)
				break
			}
		}
		if len(connList) == 0 {
			delete(cm.UidConnMap, conn.uid)
		}
	}
	if connList, ok := cm.GroupConnMap[conn.groupId]; ok {
		for i, c := range connList {
			if c.cid == conn.cid {
				index := len(connList) - 1
				if i+1 < len(connList)-1 {
					index = i + 1
				}
				connList = append(connList[:i], connList[index:]...)
				break
			}
		}
		if len(connList) == 0 {
			delete(cm.GroupConnMap, conn.groupId)
		}
	}
	cm.All.Remove(conn.element)
}

func SetSendPongFunc(f func(conn *Conn, data string)) {
	sendPong = f
}

func SetSendPingFunc(f func(conn *Conn)) {
	sendPing = f
}

func SetRouterMgr(r RouterMgr) {
	routerMgr = r
}

func GetRouterMgr() RouterMgr {
	return routerMgr
}

func SetCallOnConnStateChange(f func(c *Conn, state ConnState, reason string)) {
	callOnConnStateChange = f
}

func SetDefaultConnMgr(mgr ConnMgr) {
	connMgr = mgr
}
