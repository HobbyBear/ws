package ws

import (
	"container/list"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

func InitWs(addr string, options ...Option) *Server {
	s := &Server{
		Serv:     nil,
		Upgrader: &websocket.Upgrader{},
		Logger:   defaultLogger,
		wg:       sync.WaitGroup{},
		conTicker: &ConnTick{
			mux:              sync.Mutex{},
			TickList:         list.New(),
			TickMap:          map[string]*Conn{},
			WheelIntervalSec: 60,
			TickExpireSec:    5 * 60,
		},
	}
	s.Serv = &http.Server{
		Addr:    addr,
		Handler: &mux{server: s},
	}
	for _, op := range options {
		op(s)
	}
	//s.conTicker.Start()
	return s
}
