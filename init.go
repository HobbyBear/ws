package ws

import (
	"container/list"
	"log"
	"runtime"
	"sync"
	"ws/internal/netpoll"
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
		Addr: addr,
	}
	for _, op := range options {
		op(s)
	}
	s.PollList = make([]netpoll.Poller, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		s.PollList[i], _ = netpoll.New(&netpoll.Config{
			OnWaitError: func(err error) {
				log.Println(err)
			},
		})
	}
	s.conTicker.Start()
	return s
}
