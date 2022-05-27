package ws

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/panjf2000/ants"
	"github.com/pborman/uuid"
	"go.uber.org/atomic"
	"log"
	"net"
	"sync"
	"time"
	"ws/netpoll"
)

var (
	analyzeProtocolPool, _ = ants.NewPool(128)
	handlePool, _          = ants.NewPool(128)
	acceptPool, _          = ants.NewPool(128)
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
	Seq                 atomic.Int32
	Addr                string
	UpgradeDeadline     time.Duration // 升级ws协议的超时时间
	ReadHeaderDeadline  time.Duration // 读取header的超时时间
	ReadPayloadDeadline time.Duration // 读取payload的超时时间
	cpuNum              int
	connMgr             ConnMgr
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
	var err error
	s.Listener, err = net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		rawConn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("Listener.Accept err=%s \n", err)
			return
		}
		acceptPool.Submit(func() {
			rawConn.SetReadDeadline(time.Now().Add(s.UpgradeDeadline))
			_, err = ws.Upgrade(rawConn)
			if err != nil {
				Errorf("握手失败 err=%s", err)
				rawConn.Close()
				return
			}
			conn := &Conn{
				cid:             uuid.New(),
				uid:             "",
				mux:             sync.Mutex{},
				rawConn:         rawConn,
				stopSig:         atomic.Int32{},
				server:          s,
				groupId:         "",
				lastReceiveTime: time.Now(),
				element:         nil,
				tickElement:     nil,
				topic:           "",
			}
			s.handleConn(conn)
		})
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
	s.Listener.Close()

	closewait := sync.WaitGroup{}
	for _, conn := range s.connMgr.GetAllConn() {
		closewait.Add(1)
		go func() {
			conn.Close("服务器关闭")
			closewait.Done()
		}()
	}
	closewait.Wait()
	for _, poll := range s.PollList {
		poll.Close()
	}
	s.wg.Wait()
}

type loadBalancePolicy int

const (
	roundRobinLoadBalance loadBalancePolicy = iota
)

func (s *Server) selectPoller(policy loadBalancePolicy) netpoll.Poller {
	switch policy {
	case roundRobinLoadBalance:
		s.Seq.Inc()
		return s.PollList[int(s.Seq.Load())%(s.cpuNum-1)]
	}
	s.Seq.Inc()
	return s.PollList[int(s.Seq.Load())%(s.cpuNum-1)]
}

func (s *Server) addConn(conn *Conn) {
	s.connMgr.Add(conn)
	callOnConnStateChange(conn, StateNew, "")
	s.ConnNum.Add(1)
}

func (s *Server) handleConn(conn *Conn) {
	s.addConn(conn)
	desc := netpoll.Must(netpoll.Handle(conn.rawConn, netpoll.EventRead))
	poller := s.selectPoller(roundRobinLoadBalance)
	conn.poll = poller
	conn.pollDesc = desc

	err := poller.Start(desc, func(event netpoll.Event, notice *sync.WaitGroup) {
		callOnConnStateChange(conn, StateActive, "")
		if event&netpoll.EventRead == 0 {
			return
		}
		if event&netpoll.EventHup != 0 || event&netpoll.EventReadHup != 0 {
			conn.Close("正常关闭")
			return
		}
		analyzeProtocolPool.Submit(func() {
			defer notice.Done()
			packets, err := protocolContent(conn, s.ReadHeaderDeadline, s.ReadPayloadDeadline)
			for _, msg := range packets {
				packet := msg
				handlePool.Submit(func() {
					s.wg.Add(1)
					defer s.wg.Done()
					dataHandler(conn, packet.data, packet.header.OpCode)
				})
			}
			if err != nil {
				conn.Close("读取请求体" + err.Error())
				return
			}
		})

	})
	if err != nil {
		Errorf("poll start fail err=%s", err)
	}
}
