package ws

import (
	"container/list"
	"sync"
	"time"
)

type ConnTick struct {
	mux              sync.Mutex
	TickList         *list.List
	TickMap          map[string]*Conn
	WheelIntervalSec int // 轮询心跳链表的间隔
	TickExpireSec    int // 心跳过期的阀值
}

func (c *ConnTick) AddTickConn(conn *Conn) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if _, ok := c.TickMap[conn.cid]; ok {
		delete(c.TickMap, conn.cid)
		c.TickList.Remove(conn.tickElement)
	}
	conn.lastReceiveTime = time.Now()
	conn.tickElement = c.TickList.PushBack(conn)
	c.TickMap[conn.cid] = conn
}

func (c *ConnTick) Start() {
	go func() {
		timer := time.NewTimer(time.Duration(c.WheelIntervalSec) * time.Second)
		for {
			select {
			case <-timer.C:
				c.mux.Lock()
				for e := c.TickList.Front(); e != nil; {
					if conn, ok := e.Value.(*Conn); ok && time.Now().Sub(conn.lastReceiveTime) > time.Duration(c.TickExpireSec)*time.Second {
						delete(c.TickMap, conn.cid)
						next := e.Next()
						go func() {
							conn.Close("超时关闭")
						}()
						c.TickList.Remove(conn.tickElement)
						e = next
					} else {
						break
					}
				}
				c.mux.Unlock()
				timer.Reset(time.Duration(c.WheelIntervalSec) * time.Second)
			}
		}
	}()
}
