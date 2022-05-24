package ws

import (
	"container/list"
	"easygo/netpoll"
	"log"
	"sync"
)

func InitWs(addr string, options ...Option) *Server {
	s := &Server{
		Logger: defaultLogger,
		wg:     sync.WaitGroup{},
		conTicker: &ConnTick{
			mux:              sync.Mutex{},
			TickList:         list.New(),
			TickMap:          map[string]*Conn{},
			WheelIntervalSec: 60,
			TickExpireSec:    5 * 60,
		},
	}
	for _, op := range options {
		op(s)
	}
	s.Poll, _ = netpoll.New(&netpoll.Config{
		OnWaitError: func(err error) {
			log.Println(err)
		},
	})
	//s.conTicker.Start()
	return s
}
