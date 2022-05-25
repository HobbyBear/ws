package ws

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants"
	"github.com/pborman/uuid"
	"go.uber.org/atomic"
	"log"
	"net"
	"sync"
	"time"
	"ws/internal/netpoll"
)

var (
	analyzeProtocolPool, _ = ants.NewPool(128)
	handlePool, _          = ants.NewPool(128)
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
	Logger              Log
	wg                  sync.WaitGroup
	ConnNum             atomic.Int32
	conTicker           *ConnTick
	Listener            net.Listener
	PollList            []netpoll.Poller
	Seq                 int
	Addr                string
	UpgradeDeadline     time.Duration // 升级ws协议的超时时间
	ReadHeaderDeadline  time.Duration // 读取header的超时时间
	ReadPayloadDeadline time.Duration // 读取payload的超时时间
}

type Option func(s *Server)

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

func (s *Server) startListen() {
	s.Listener, _ = net.Listen("tcp", s.Addr)
	for {
		rawConn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("Listener.Accept err=%s \n", err)
			return
		}
		rawConn.SetReadDeadline(time.Now().Add(s.UpgradeDeadline))
		_, err = ws.Upgrade(rawConn)
		if err != nil {
			// handle error
			Errorf("握手失败 err=%s", err)
			rawConn.Close()
			continue
		}
		conn := &Conn{
			Cid:             uuid.New(),
			Uid:             "",
			mux:             sync.Mutex{},
			rawConn:         rawConn,
			stopSig:         atomic.Int32{},
			stop:            make(chan int, 1),
			server:          s,
			GroupId:         "",
			lastReceiveTime: time.Now(),
			element:         nil,
			tickElement:     nil,
			topic:           "",
		}

		connMgr.Add(conn)
		callOnConnStateChange(conn, StateNew, "")
		s.ConnNum.Add(1)

		s.handleConn(conn)

	}
}

func (s *Server) Start(block bool) {
	go func() {
		timer := time.NewTimer(3 * time.Second)
		for {
			select {
			case <-timer.C:
				fmt.Println("连接数", s.ConnNum.Load())
				timer.Reset(3 * time.Second)
			}
		}
	}()
	if block {
		s.startListen()
		return
	}
	go s.startListen()
}

func (s *Server) ShutDown() {
	// 关闭listener
	s.Listener.Close()
	// 发送关闭帧
	allConn := connMgr.GetAllConn()
	for _, conn := range allConn {
		analyzeProtocolPool.Submit(func() {
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
	for _, poll := range s.PollList {
		poll.Close()
	}
}

func (s *Server) handleConn(conn *Conn) {
	rawConn := conn.rawConn
	desc := netpoll.Must(netpoll.Handle(rawConn, netpoll.EventRead))
	poller := s.PollList[s.Seq%7]
	s.Seq++
	conn.poll = poller
	conn.pollDesc = desc
	err := poller.Start(desc, func(event netpoll.Event) {
		callOnConnStateChange(conn, StateActive, "")
		if event&netpoll.EventRead == 0 {
			return
		}
		header, payload, err := getProtocolContent(rawConn, s.ReadHeaderDeadline, s.ReadPayloadDeadline)
		if err != nil {
			conn.Close(err.Error() + " 读取请求体")
			return
		}
		handlePool.Submit(func() {
			s.conTicker.AddTickConn(conn)
			if header.OpCode == ws.OpClose {
				conn.Close("对端主动关闭")
				return
			}
			if header.OpCode == ws.OpPing {
				sendPing(conn)
				return
			}
			if header.Masked {
				ws.Cipher(payload, header.Mask, 0)
			}
			if len(payload) == 0 {
				return
			}
			dataHandler(conn, payload, header.OpCode)
		})
	})
	if err != nil {
		Errorf("poll start fail err=%s", err)
	}
}
