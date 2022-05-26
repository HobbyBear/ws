package ws

import (
	"container/list"
	"sync"
	"testing"
	"time"
)

func TestConnTick_Start(t *testing.T) {
	tiker := &ConnTick{
		mux:      sync.Mutex{},
		TickList: list.New(),
		TickMap:  map[string]*Conn{},
	}
	tiker.AddTickConn(&Conn{
		cid:             "123",
		uid:             "",
		lastReceiveTime: time.Time{},
	})
	tiker.Start()
	time.Sleep(time.Hour)
}
