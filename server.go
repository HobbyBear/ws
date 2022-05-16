package ws

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	// 换成自定义的是否会更快，因为这个库的线程池有些功能用不到
	p, _ = ants.NewPool(10)
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
	wg           sync.WaitGroup
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

func (s *Server) ShutDown() {
	// 关闭listener
	s.Serv.Shutdown(context.TODO())
	// 发送关闭帧
	allConn := connMgr.GetAllConn()
	for _, conn := range allConn {
		p.Submit(func() {
			conn.WriteMsg(&RawMsg{WsMsgType: websocket.CloseMessage, DeadLine: time.Now().Add(time.Second)})
		})
	}
	// 对已经收到的帧进行处理
	s.wg.Wait()
	time.Sleep(500 * time.Millisecond)
	allConn = connMgr.GetAllConn()
	for _, conn := range allConn {
		conn.Close()
	}
}

func (s *Server) handleConn(conn *Conn) {
	callOnConnStateChange(conn, StateActive)
	go func() {
		for {
			mt, message, err := conn.wsConn.ReadMessage()
			// todo 应该对错误进行判断
			if err != nil {
				s.Logger.Debugf("conn read err=%s cid=%s", err, conn.Cid)
				conn.Close()
				return
			}
			err = p.Submit(func() {
				s.wg.Add(1)
				defer s.wg.Done()
				dataHandler(conn, message, mt)
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		for msg := range conn.sendBuffer {
			if msg.WsMsgType == 0 {
				return
			}
			err := msg.write(conn)
			if err != nil {
				log.Println("write:", err, string(msg.Data), msg.WsMsgType)
				return
			}
		}
	}()
}
