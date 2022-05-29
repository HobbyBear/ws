package wsgateway

import (
	"github.com/gobwas/ws"
	"sync"
	"ws/msg"
)

// 一些预置的消息类型
const (
	SubMsgType   = "sub_topic" // 订阅的消息类型
	Login        = "login"
	UnSubMsgType = "unsub_topic" // 取消订阅消息
	Ping         = "ping"
	Close        = "close"
	Pong         = "pong"
)

type WServerMux struct {
	mu sync.RWMutex
	m  map[string]Handler
}

func NewWsMux() *WServerMux {
	mux := &WServerMux{
		mu: sync.RWMutex{},
		m:  make(map[string]Handler),
	}
	mux.AddHandler(SubMsgType, HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
		conn.topic = reqMsg.Topic
		conn.server.CallConnStateChange(conn, StateSubTopic)
	}))
	mux.AddHandler(Login, HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
		conn.uid = reqMsg.Uid
		conn.server.CallConnStateChange(conn, StateLogin)
	}))
	mux.AddHandler(UnSubMsgType, HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
		conn.topic = ""
		conn.server.CallConnStateChange(conn, StateUnSubTopic)
	}))
	mux.AddHandler(Ping, HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
		conn.WriteMsg(ws.OpPong, nil)
	}))
	mux.AddHandler(Pong, HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
		return
	}))
	mux.AddHandler(Close, HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
		return
	}))
	return mux
}

func (w *WServerMux) AddHandler(path string, handler Handler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.m[path] = handler
}

func (w *WServerMux) ServerWs(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if handler, ok := w.m[reqMsg.Path]; ok {
		handler.ServerWs(opCode, reqMsg, conn)
	}
}

type Handler interface {
	ServerWs(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn)
}

var DefaultMux = NewWsMux()

type HandlerFunc func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn)

func (h HandlerFunc) ServerWs(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *Conn) {
	h(opCode, reqMsg, conn)
}

func AddHandler(path string, h Handler) {
	DefaultMux.AddHandler(path, h)
}
