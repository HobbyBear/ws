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
	GroupId         string
	lastReceiveTime time.Time
	element         *list.Element
	tickElement     *list.Element
	topic           string
}

type RawMsg struct {
	WsMsgType int       `json:"-"`
	MsgType   string    `json:"msg_id,omitempty"`
	Content   []byte    `json:"content,omitempty"`
	DeadLine  time.Time `json:"-"`
}

func (c *Conn) WriteMsg(data *RawMsg) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.stopSig.Load() == 1 {
		return nil
	}
	if isControl(data.WsMsgType) {
		return c.wsConn.WriteControl(data.WsMsgType, data.Content, data.DeadLine)
	}
	if isData(data.WsMsgType) {
		return c.wsConn.WriteMessage(data.WsMsgType, data.Content)
	}
	return errors.New("websocket: bad write message type")
}

func (c *Conn) Close(reason string) {
	if c.stopSig.Load() == 1 {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.stopSig.Store(1)
	c.server.ConnNum.Sub(1)
	defaultConnMgr.Del(c)
	callOnConnStateChange(c, StateClosed, reason)
	c.stop <- 1
	c.wsConn.Close()
}

func isControl(frameType int) bool {
	return frameType == websocket.CloseMessage || frameType == websocket.PingMessage || frameType == websocket.PongMessage
}

func isData(frameType int) bool {
	return frameType == websocket.TextMessage || frameType == websocket.BinaryMessage
}
