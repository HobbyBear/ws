package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants"
	"go.uber.org/atomic"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	// 换成自定义的是否会更快，因为这个库的线程池有些功能用不到
	p, _ = ants.NewPool(1000)
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
	Serv      *http.Server
	Upgrader  *websocket.Upgrader
	Logger    Log
	wg        sync.WaitGroup
	ConnNum   atomic.Int32
	conTicker *ConnTick
}

type Option func(s *Server)

func SetUpgrader(upgrader *websocket.Upgrader) Option {
	return func(s *Server) {
		s.Upgrader = upgrader
	}
}

func SetLogger(logger Log) Option {
	return func(s *Server) {
		s.Logger = logger
	}
}

func SetTickInterval(intervalSec int) Option {
	return func(s *Server) {
		s.conTicker.WheelIntervalSec = intervalSec
	}
}

func SetTickExpire(expireSec int) Option {
	return func(s *Server) {
		s.conTicker.TickExpireSec = expireSec
	}
}

func (s *Server) Start() {
	go s.Serv.ListenAndServe()
	go func() {
		timer := time.NewTimer(3 * time.Second)
		for {
			select {
			case <-timer.C:
				fmt.Println("连接数", s.ConnNum.Load())
				if count.Load() != 0 {
					fmt.Println("消耗时间", total.Load()/count.Load())
				}
				timer.Reset(3 * time.Second)
			}
		}
	}()
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
		conn.Close("服务器关闭")
	}
}

var (
	total = atomic.NewInt64(0)
	count = atomic.NewInt64(0)
)

func (s *Server) handleConn(conn *Conn) {
	callOnConnStateChange(conn, StateActive, "")
	go func() {
		for {
			start := time.Now()
			mt, message, err := conn.wsConn.ReadMessage()
			if err != nil {
				conn.Close("读取消息失败" + err.Error())
				return
			}
			count.Inc()
			total.Add(time.Now().Sub(start).Milliseconds())
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
}
