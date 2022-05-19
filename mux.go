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
		GroupId:         "",
		lastReceiveTime: time.Now(),
		element:         nil,
		tickElement:     nil,
		topic:           "",
	}

	connMgr.Add(conn)
	callOnConnStateChange(conn, StateNew, "")
	m.server.ConnNum.Add(1)
	conn.wsConn.SetPingHandler(func(message string) error {
		m.server.conTicker.AddTickConn(conn)
		sendPong(conn, message)
		return nil
	})
	m.server.handleConn(conn)
}
