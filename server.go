package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type ConnState int

func (c ConnState) String() string {
	switch c {
	case StateNew:
		return "new"
	case StateActive:
		return "active"
	case StateClosed:
		return "closed"
	}
	return ""
}

const (
	StateNew ConnState = iota

	StateActive

	StateClosed
)

type Request struct {
	Conn    *Conn
	Data    []byte
	MsgType int
}

type Server struct {
	Serv         *http.Server
	Upgrader     *websocket.Upgrader
	Logger       Log
	PingInterval time.Duration
}

type Option func(s *Server)

func Upgrader(upgrader *websocket.Upgrader) Option {
	return func(s *Server) {
		s.Upgrader = upgrader
	}
}

func Logger(logger Log) Option {
	return func(s *Server) {
		s.Logger = logger
	}
}

func (s *Server) Start() {
	go s.Serv.ListenAndServe()
}

// todo
func (s *Server) ShutDown() {

}

func (s *Server) handleConn(conn *Conn) {
	callOnConnStateChange(conn, StateActive)
	go func() {
		for {
			mt, message, err := conn.wsConn.ReadMessage()
			if err != nil {
				s.Logger.Debugf("conn read err=%s cid=%s", err, conn.Cid)
				conn.Close()
				return
			}
			// todo 携程池来做
			go requestHandler(&Request{Conn: conn, Data: message, MsgType: mt})
			if err != nil {
				Errorf("handle Conn err=%s cid=%s", err, conn.Cid)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-conn.stop:
				return
			case c := <-conn.sendBuffer:
				if c.MsgType == 0 {
					return
				}
				err := conn.wsConn.WriteMessage(c.MsgType, c.Data)
				if err != nil {
					log.Println("write:", err, string(c.Data), c.MsgType)
					return
				}
			}
		}
	}()
	<-conn.stop
}
