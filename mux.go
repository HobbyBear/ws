package ws

import (
	"github.com/pborman/uuid"
	"go.uber.org/atomic"
	"log"
	"net/http"
	"sync"
	"time"
)

type mux struct {
	server *Server
}

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := m.server.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	conn := &Conn{
		Cid:             uuid.New(),
		Uid:             "",
		mux:             sync.Mutex{},
		wsConn:          c,
		stopSig:         atomic.Int32{},
		stop:            make(chan int, 1),
		server:          m.server,
		sendBuffer:      make(chan *RawMsg, 100),
		GroupId:         "",
		pingTimer:       time.NewTimer(m.server.PingInterval),
		lastReceiveTime: time.Now(),
	}

	connMgr.Add(conn)
	callOnConnStateChange(conn, StateNew)
	conn.wsConn.SetPingHandler(func(message string) error {
		conn.lastReceiveTime = time.Now()
		sendPong(conn, message)
		return nil
	})
	conn.wsConn.SetPongHandler(func(message string) error {
		conn.lastReceiveTime = time.Now()
		return nil
	})
	conn.KeepAlive()
	m.server.handleConn(conn)
	//if conn.Auth() {
	//	m.server.handleConn(conn)
	//} else {
	//	conn.Close()
	//}
}
