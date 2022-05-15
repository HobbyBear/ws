package ws

import (
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"log"
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
	sendBuffer      chan Msg
	readBuffer      chan Msg
	GroupId         string
	pingTimer       *time.Timer
	lastReceiveTime time.Time
}

type Msg struct {
	MsgType int
	Data    []byte
}

func (c *Conn) WriteMsg(data Msg) {
	c.sendBuffer <- data

}

func (c *Conn) ReadMsg() Msg {
	return <-c.readBuffer
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
					pingHandler(c)
					Infof("send ping msg cid=%s", c.Cid)
				}
			case <-c.stop:
				c.pingTimer.Stop()
				return
			}
		}
	}()
}

func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	c.pingTimer.Reset(c.server.PingInterval)
	c.lastReceiveTime = time.Now()
	return c.wsConn.ReadMessage()
}

func (c *Conn) ReadJson(v interface{}) (err error) {
	c.pingTimer.Reset(c.server.PingInterval)
	c.lastReceiveTime = time.Now()
	return c.wsConn.ReadJSON(v)
}

func (c *Conn) Auth() bool {
	_, message, _ := c.ReadMessage()
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
	close(c.readBuffer)
	c.wsConn.Close()
}

func (c *Conn) ShutDown() {
	if c.stopSig.Load() == 1 {
		return
	}
	c.stopSig.Store(1)
	c.mux.Lock()
	defer c.mux.Unlock()
	defaultConnMgr.Del(c)
	callOnConnStateChange(c, StateClosed)
	c.stop <- 1
	close(c.sendBuffer)
	close(c.readBuffer)
	for msg := range c.sendBuffer {
		err := c.wsConn.WriteMessage(msg.MsgType, msg.Data)
		if err != nil {
			log.Println(err)
			break
		}
	}
	c.wsConn.Close()
}
