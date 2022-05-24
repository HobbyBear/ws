package ws

import (
	"bufio"
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
	"ws/internal/trylock"
)

var (
	analyzeProtocolPool, _ = ants.NewPool(1000)
	handlePool, _          = ants.NewPool(1000)
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
	Logger    Log
	wg        sync.WaitGroup
	ConnNum   atomic.Int32
	conTicker *ConnTick
	Listener  net.Listener
	PollList  []netpoll.Poller
	Seq       int
	Addr      string
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
		// todo 超时控制，关闭掉慢客户端

		_, err = ws.Upgrade(rawConn)
		if err != nil {
			// handle error
			log.Fatal(err)
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
			reader:          bufio.NewReader(rawConn),
			writer:          bufio.NewWriter(rawConn),
			protocolLock:    trylock.NewMutex(),
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
	poller.Start(desc, func(event netpoll.Event) {
		callOnConnStateChange(conn, StateActive, "")
		analyzeProtocolPool.Submit(func() {
			if !conn.protocolLock.TryLock() {
				return
			}
			defer conn.protocolLock.Unlock()

			if event&netpoll.EventRead == 0 {
				return
			}
			header, payload, err := getProtocolContent(conn.reader)
			if err != nil {
				conn.Close(err.Error())
				return
			}
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
			handlePool.Submit(func() {
				dataHandler(conn, payload, header.OpCode)
			})
		})
	})
}
