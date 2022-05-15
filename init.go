package ws

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

func InitWs(addr string, options ...Option) *Server {
	s := &Server{
		Serv:         nil,
		PingInterval: 3 * time.Second,
		Upgrader:     &websocket.Upgrader{},
		Logger:       defaultLogger,
	}
	s.Serv = &http.Server{
		Addr:    addr,
		Handler: &mux{server: s},
	}
	for _, op := range options {
		op(s)
	}
	return s
}
