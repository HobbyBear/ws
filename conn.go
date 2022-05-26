package ws

import (
	"container/list"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"log"
	"net"
	"sync"
	"time"
	"ws/netpoll"
)

type Conn struct {
	cid             string
	uid             string
	mux             sync.Mutex
	rawConn         net.Conn
	stopSig         atomic.Int32
	server          *Server
	groupId         string
	lastReceiveTime time.Time
	element         *list.Element
	tickElement     *list.Element
	topic           string
	poll            netpoll.Poller
	pollDesc        *netpoll.Desc
}

type RawMsg struct {
	WsMsgType ws.OpCode `json:"-"`
	MsgType   string    `json:"msg_id,omitempty"`
	Content   []byte    `json:"content,omitempty"`
	DeadLine  time.Time `json:"-"`
}

// todo 异步写
func (c *Conn) WriteMsg(data *RawMsg) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.stopSig.Load() == 1 {
		return nil
	}
	if data.WsMsgType.IsControl() {
		w := newBuffWriter(c.rawConn)
		wsutil.WriteServerMessage(w, data.WsMsgType, data.Content)
		putBuffWriter(w)
		return nil
	}
	if data.WsMsgType.IsData() {
		w := newBuffWriter(c.rawConn)
		wsutil.WriteServerMessage(w, data.WsMsgType, data.Content)
		putBuffWriter(w)
		return nil
	}
	return errors.New("websocket: bad write message type")
}

func (c *Conn) Close(reason string) {
	if c.stopSig.Load() == 1 {
		return
	}
	log.Println(reason, "连接关闭")
	c.mux.Lock()
	defer c.mux.Unlock()
	c.stopSig.Store(1)
	c.server.ConnNum.Sub(1)
	c.server.connMgr.Del(c)
	callOnConnStateChange(c, StateClosed, reason)
	c.poll.Stop(c.pollDesc)
	c.rawConn.Close()
}
