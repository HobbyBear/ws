package wsgateway

import (
	"fmt"
	"github.com/gobwas/ws"
	jsoniter "github.com/json-iterator/go"
	"github.com/pborman/uuid"
	"go.uber.org/atomic"
	"log"
	"net"
	"sync"
	"time"
	"ws/broker"
	"ws/msg"
	"ws/pkg/logger"
	"ws/pkg/netpoll"
)

type ConnState int

func (c ConnState) String() string {
	switch c {
	case StateNew:
		return "new"
	case StateClosed:
		return "closed"
	}
	return ""
}

const (
	StateNew ConnState = iota

	StateClosed

	StateLogin

	StateEnterRoom

	StateSubTopic

	StateUnSubTopic
)

type Server struct {
	Logger              logger.Log
	wg                  sync.WaitGroup
	connNum             atomic.Int32
	conTicker           *connTick
	listener            net.Listener
	pollList            []netpoll.Poller
	pollSeq             atomic.Int32
	Addr                string
	UpgradeDeadline     time.Duration // 升级ws协议的超时时间
	ReadHeaderDeadline  time.Duration // 读取header的超时时间
	ReadPayloadDeadline time.Duration // 读取payload的超时时间
	cpuNum              int
	ConnMgr             ConnMgr
	Broker              broker.Broker
	Mux                 Handler
	TickOpen            bool // true 代表开启心跳检测
	CallConnStateChange func(c *Conn, state ConnState)
}

func (s *Server) startListen() {
	var err error
	s.listener, err = net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		rawConn, err := s.listener.Accept()
		if err != nil {
			log.Printf("listener.Accept err=%s \n", err)
			return
		}
		acceptPool.Submit(func() {
			rawConn.SetReadDeadline(time.Now().Add(s.UpgradeDeadline))
			_, err = ws.Upgrade(rawConn)
			if err != nil {
				logger.Errorf("握手失败 err=%s", err)
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
				roomId:          "",
				lastReceiveTime: time.Now(),
				allElement:      nil,
				tickElement:     nil,
				topic:           "",
			}
			s.handleConn(conn)
		})
	}
}

func (s *Server) Start(block bool) {
	// todo 调试代码，待删除
	go func() {
		timer := time.NewTimer(3 * time.Second)
		for {
			select {
			case <-timer.C:
				fmt.Println("连接数", s.connNum.Load())
				timer.Reset(3 * time.Second)
			}
		}
	}()
	s.conTicker.open = s.TickOpen
	initConsumer(s)
	s.conTicker.Start()

	if block {
		s.startListen()
		return
	}
	go s.startListen()
}

func (s *Server) ShutDown() {
	s.listener.Close()
	s.Broker.Close()
	closewait := sync.WaitGroup{}
	for _, conn := range s.ConnMgr.GetAllConn() {
		closewait.Add(1)
		go func() {
			conn.Close("服务器关闭")
			closewait.Done()
		}()
	}
	closewait.Wait()
	for _, poll := range s.pollList {
		poll.Close()
	}
	s.wg.Wait()
}

func (s *Server) onConnClose(c *Conn) {
	s.connNum.Sub(1)
	s.ConnMgr.Del(c)
	c.poll.Stop(c.pollDesc)
	s.CallConnStateChange(c, StateClosed)

}

type loadBalancePolicy int

const (
	roundRobinLoadBalance loadBalancePolicy = iota
)

func (s *Server) selectPoller(policy loadBalancePolicy) netpoll.Poller {
	switch policy {
	case roundRobinLoadBalance:
		s.pollSeq.Inc()
		return s.pollList[int(s.pollSeq.Load())%(s.cpuNum-1)]
	}
	s.pollSeq.Inc()
	return s.pollList[int(s.pollSeq.Load())%(s.cpuNum-1)]
}

func (s *Server) addConn(conn *Conn) {
	s.ConnMgr.Add(conn)
	s.CallConnStateChange(conn, StateNew)
	s.connNum.Add(1)
}

func (s *Server) handleConn(conn *Conn) {
	s.addConn(conn)
	desc := netpoll.Must(netpoll.Handle(conn.rawConn, netpoll.EventRead))
	poller := s.selectPoller(roundRobinLoadBalance)
	conn.poll = poller
	conn.pollDesc = desc

	err := poller.Start(desc, func(event netpoll.Event, notice *sync.WaitGroup) {
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
			for _, packetData := range packets {
				packet := packetData
				handlePool.Submit(func() {
					s.wg.Add(1)
					defer s.wg.Done()
					var (
						reqMsg = &msg.ReqMsg{}
						json   = jsoniter.ConfigCompatibleWithStandardLibrary
						err    error
					)
					if len(packet.data) != 0 {
						err = json.Unmarshal(packet.data, reqMsg)
						if err != nil {
							logger.Errorf("data reqMsg is invalid err=%s data=%s", err, string(reqMsg.Data))
							return
						}
					}
					switch packet.header.OpCode {
					case ws.OpPing:
						reqMsg.Path = Ping
					case ws.OpPong:
						reqMsg.Path = Pong
					case ws.OpClose:
						reqMsg.Path = Close
					}
					s.Mux.ServerWs(packet.header.OpCode, reqMsg, conn)
				})
			}
			if err != nil {
				conn.Close("读取请求体" + err.Error())
				return
			}
		})

	})
	if err != nil {
		logger.Errorf("poll start fail err=%s", err)
	}
}
