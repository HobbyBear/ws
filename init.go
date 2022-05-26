package ws

import (
	"container/list"
	"go.uber.org/atomic"
	"runtime"
	"sync"
	"time"
	"ws/netpoll"
)

func InitWs(addr string, options ...Option) *Server {
	s := &Server{
		Logger:  defaultLogger,
		wg:      sync.WaitGroup{},
		ConnNum: atomic.Int32{},
		conTicker: &ConnTick{
			mux:              sync.Mutex{},
			TickList:         list.New(),
			TickMap:          map[string]*Conn{},
			WheelIntervalSec: 20,
			TickExpireSec:    10,
		},
		Listener:            nil,
		PollList:            nil,
		Seq:                 0,
		Addr:                addr,
		UpgradeDeadline:     3 * time.Second,
		ReadHeaderDeadline:  1 * time.Second,
		ReadPayloadDeadline: 2 * time.Second,
		cpuNum:              runtime.NumCPU(),
		connMgr:             defaultConnMgr,
	}
	for _, op := range options {
		op(s)
	}
	s.PollList = make([]netpoll.Poller, s.cpuNum)
	for i := 0; i < s.cpuNum; i++ {
		s.PollList[i], _ = netpoll.New(&netpoll.Config{
			OnWaitError: func(err error) {
				Errorf("poll internal error err=%s", err)
			},
		})
	}
	//s.conTicker.Start()
	return s
}
