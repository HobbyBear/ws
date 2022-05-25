package ws

import (
	"container/list"
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
			WheelIntervalSec: 20,
			TickExpireSec:    10,
		},
		Addr: addr,
	}
	for _, op := range options {
		op(s)
	}
	numCpu := runtime.NumCPU()
	s.PollList = make([]netpoll.Poller, numCpu)
	for i := 0; i < numCpu; i++ {
		s.PollList[i], _ = netpoll.New(&netpoll.Config{
			OnWaitError: func(err error) {
				Errorf("poll internal error err=%s", err)
			},
		})
	}
	s.conTicker.Start()
	return s
}
