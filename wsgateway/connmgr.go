package wsgateway

import (
	"container/list"
	"sync"
)

var (
	defaultConnMgr = &DefaultConnMgr{
		UidConnMap: make(map[string][]*Conn),
		RoomsMap:   make(map[string]*Room),
		All:        list.New(),
		mux:        sync.Mutex{},
	}
)

type ConnMgr interface {
	Del(c *Conn)
	Add(c *Conn)
	GetConnByUids(uids []string) map[string][]*Conn
	GetConnByRoomId(roomId string) []*Conn
	GetAllConn() []*Conn
	GetRoom() []*Room
}

type DefaultConnMgr struct {
	UidConnMap map[string][]*Conn
	RoomsMap   map[string]*Room
	All        *list.List
	mux        sync.Mutex
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
	if len(c.roomId) != 0 {
		room, ok := cm.RoomsMap[c.roomId]
		if !ok {
			room = NewRoom(c.roomId)
			cm.RoomsMap[c.roomId] = room
		}
		room.AddCoon(c)
	}
	e := cm.All.PushFront(c)
	c.allElement = e
}

func (cm *DefaultConnMgr) GetConnByUids(uids []string) map[string][]*Conn {
	data := make(map[string][]*Conn)
	for _, uid := range uids {
		data[uid] = cm.UidConnMap[uid]
	}
	return data
}

func (cm *DefaultConnMgr) GetConnByRoomId(roomId string) []*Conn {
	room, ok := cm.RoomsMap[roomId]
	if !ok {
		return nil
	}
	return room.GetConnList()
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
	if room, ok := cm.RoomsMap[conn.roomId]; ok {
		left := room.Del(conn)
		if left == 0 {
			delete(cm.RoomsMap, conn.roomId)
		}
	}
	cm.All.Remove(conn.allElement)
}

func (cm *DefaultConnMgr) GetRoom() []*Room {
	roomList := make([]*Room, 0)
	for _, room := range cm.RoomsMap {
		roomList = append(roomList, room)
	}
	return roomList
}
