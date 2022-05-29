package wsgateway

import (
	"container/list"
	"sync"
)

type Room struct {
	rid      string
	rLock    sync.RWMutex
	connList *list.List
}

func NewRoom(rid string) *Room {
	return &Room{
		rid:      rid,
		rLock:    sync.RWMutex{},
		connList: list.New(),
	}
}

func (r *Room) AddCoon(c *Conn) {
	r.rLock.Lock()
	defer r.rLock.Unlock()
	r.connList.PushFront(c)
}

func (r *Room) GetConnList() []*Conn {
	r.rLock.RLock()
	defer r.rLock.Unlock()
	connList := make([]*Conn, 0)
	for l := r.connList.Front(); l != nil; l = l.Next() {
		connList = append(connList, l.Value.(*Conn))
	}
	return connList
}

func (r *Room) Del(c *Conn) int {
	r.rLock.Lock()
	defer r.rLock.Unlock()
	for e := r.connList.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*Conn)
		if conn.cid == c.cid {
			r.connList.Remove(e)
		}
	}
	return r.connList.Len()
}
