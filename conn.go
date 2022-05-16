package ws

import (
	"container/list"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"sync"
	"time"
)

type Conn struct {
	Cid             string
	Uid             string
	mux             sync.Mutex
	wsConn          *websocket.Conn
	stopSig         atomic.Int32
	stop            chan int
	server          *Server
	readBuffer      chan RawMsg
	sendBuffer      chan *RawMsg
	GroupId         string
	pingTimer       *time.Timer
	lastReceiveTime time.Time
	element         *list.Element
}

type RawMsg struct {
	WsMsgType int
	Data      []byte
	DeadLine  time.Time
}

func (m *RawMsg) write(conn *Conn) error {
	if isControl(m.WsMsgType) {
		return conn.wsConn.WriteControl(m.WsMsgType, m.Data, m.DeadLine)
	}
	if isData(m.WsMsgType) {
		return conn.wsConn.WriteMessage(m.WsMsgType, m.Data)
	}
	return errors.New("websocket: bad write message type")
}

func (c *Conn) WriteMsg(data *RawMsg) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.stopSig.Load() == 1 {
		return
	}
	c.sendBuffer <- data
}

func (c *Conn) KeepAlive() {
	go func() {
		for {
			c.pingTimer.Reset(c.server.PingInterval)
			select {
			case <-c.pingTimer.C:
				if time.Now().Sub(c.lastReceiveTime) > 10*time.Second {
					c.Close()
				} else {
					sendPing(c)
				}
			case <-c.stop:
				c.pingTimer.Stop()
				return
			}
		}
	}()
}

func (c *Conn) Auth() bool {
	_, message, _ := c.wsConn.ReadMessage()
	authMsg, success := authHandler(message)
	if !success {
		return false
	}
	c.Uid = authMsg.Uid
	c.GroupId = authMsg.GroupId
	return true
}

func (c *Conn) Close() {
	if c.stopSig.Load() == 1 {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.stopSig.Store(1)
	defaultConnMgr.Del(c)
	callOnConnStateChange(c, StateClosed)
	c.stop <- 1
	close(c.sendBuffer)
	c.wsConn.Close()
}

func isControl(frameType int) bool {
	return frameType == websocket.CloseMessage || frameType == websocket.PingMessage || frameType == websocket.PongMessage
}

func isData(frameType int) bool {
	return frameType == websocket.TextMessage || frameType == websocket.BinaryMessage
}
