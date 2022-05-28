package wsgateway

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
	"ws/pkg/bufferpool"
	"ws/pkg/netpoll"
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

// todo 异步写
func (c *Conn) WriteMsg(opCode ws.OpCode, data []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.stopSig.Load() == 1 {
		return nil
	}
	if opCode.IsControl() {
		w := bufferpool.NewBuffWriter(c.rawConn)
		wsutil.WriteServerMessage(w, opCode, data)
		bufferpool.PutBuffWriter(w)
		return nil
	}
	if opCode.IsData() {
		w := bufferpool.NewBuffWriter(c.rawConn)
		wsutil.WriteServerMessage(w, opCode, data)
		bufferpool.PutBuffWriter(w)
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
	c.server.onConnClose(c)
	c.rawConn.Close()
}
