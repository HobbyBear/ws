package wsgateway

import (
	"container/list"
	"sync"
)

var (
	defaultConnMgr = &DefaultConnMgr{
		UidConnMap:   make(map[string][]*Conn),
		GroupConnMap: make(map[string][]*Conn),
		All:          list.New(),
		mux:          sync.Mutex{},
	}
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
