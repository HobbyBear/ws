package ws

import (
	"container/list"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"net"
	"sync"
	"time"
)

type Conn struct {
	Cid             string
	Uid             string
	mux             sync.Mutex
	rawConn         net.Conn
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
	WsMsgType ws.OpCode `json:"-"`
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
		return wsutil.WriteServerMessage(c.rawConn, data.WsMsgType, data.Content)
	}
	if isData(data.WsMsgType) {
		return wsutil.WriteServerMessage(c.rawConn, data.WsMsgType, data.Content)
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
	c.rawConn.Close()
}

func isControl(frameType ws.OpCode) bool {
	return frameType == ws.OpClose || frameType == ws.OpPing || frameType == ws.OpPong
}

func isData(frameType ws.OpCode) bool {
	return frameType == ws.OpText || frameType == ws.OpBinary
}
