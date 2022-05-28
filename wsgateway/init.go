package wsgateway

import (
	"container/list"
	"go.uber.org/atomic"
	"runtime"
	"sync"
	"time"
	"ws/pkg/logger"
	"ws/pkg/netpoll"
)

func Init(addr string, options ...Option) *Server {
	s := &Server{
		Logger:  logger.Default,
		wg:      sync.WaitGroup{},
		connNum: atomic.Int32{},
		conTicker: &connTick{
			mux:              sync.Mutex{},
			TickList:         list.New(),
			TickMap:          map[string]*Conn{},
			WheelIntervalSec: 20,
			TickExpireSec:    10,
		},
		listener:            nil,
		pollList:            nil,
		pollSeq:             atomic.Int32{},
		Addr:                addr,
		UpgradeDeadline:     3 * time.Second,
		ReadHeaderDeadline:  1 * time.Second,
		ReadPayloadDeadline: 2 * time.Second,
		cpuNum:              runtime.NumCPU(),
		ConnMgr:             defaultConnMgr,
		Mux:                 DefaultMux,
		CallConnStateChange: func(c *Conn, state ConnState) {},
	}
	for _, option := range options {
		option(s)
	}

	s.pollList = make([]netpoll.Poller, s.cpuNum)
	for i := 0; i < s.cpuNum; i++ {
		s.pollList[i], _ = netpoll.New(&netpoll.Config{
			OnWaitError: func(err error) {
				logger.Errorf("poll internal error err=%s", err)
			},
		})
	}
	return s
}
