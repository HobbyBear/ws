package ws

import (
	"bufio"
	"easygo/netpoll"
	"encoding/binary"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants"
	"github.com/pborman/uuid"
	"go.uber.org/atomic"
	"io"
	"log"
	"net"
	"sync"
	"time"
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
	s.Listener, _ = net.Listen("tcp", "localhost:8080")
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

func (s *Server) Start() {
	go func() {
		timer := time.NewTimer(3 * time.Second)
		for {
			select {
			case <-timer.C:
				fmt.Println("连接数", s.ConnNum.Load())
				if count.Load() != 0 {
					fmt.Println("解析协议时间", total.Load()/count.Load())
				}
				timer.Reset(3 * time.Second)
			}
		}
	}()
	go s.startListen()
}

func (s *Server) ShutDown() {
	// 关闭listener
	// todo poll 关闭
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
}

var (
	total = atomic.NewInt64(0)
	count = atomic.NewInt64(0)
)

func (s *Server) handleConn(conn *Conn) {
	rawConn := conn.rawConn
	desc := netpoll.Must(netpoll.Handle(rawConn, netpoll.EventRead))
	poller := s.PollList[s.Seq%7]
	s.Seq++
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
			start := time.Now()
			header, err := ReadHeader(conn.reader)
			if err != nil {
				// handle error
				s.ConnNum.Dec()
				rawConn.Close()
				poller.Stop(desc)
				return
			}

			payload, err := io.ReadAll(io.LimitReader(conn.reader, header.Length))
			count.Inc()
			total.Add(time.Now().Sub(start).Milliseconds())
			if err != nil {
				// handle error
				s.ConnNum.Dec()
				rawConn.Close()
				poller.Stop(desc)
				return
			}
			if header.OpCode == ws.OpClose {
				s.ConnNum.Dec()
				rawConn.Close()
				poller.Stop(desc)
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

const (
	bit0 = 0x80
)

// ReadHeader reads a frame header from r.
func ReadHeader(r *bufio.Reader) (h ws.Header, err error) {

	// Make slice of bytes with capacity 12 that could hold any header.
	//
	// The maximum header size is 14, but due to the 2 hop reads,
	// after first hop that reads first 2 constant bytes, we could reuse 2 bytes.
	// So 14 - 2 = 12.

	// Prepare to hold first 2 bytes to choose size of next read.
	bts, err := io.ReadAll(io.LimitReader(r, 2))
	if err != nil || len(bts) == 0 {
		return
	}

	h.Fin = bts[0]&bit0 != 0
	h.Rsv = (bts[0] & 0x70) >> 4
	h.OpCode = ws.OpCode(bts[0] & 0x0f)

	var extra int

	if bts[1]&bit0 != 0 {
		h.Masked = true
		extra += 4
	}

	length := bts[1] & 0x7f
	switch {
	case length < 126:
		h.Length = int64(length)

	case length == 126:
		extra += 2

	case length == 127:
		extra += 8

	default:
		err = ws.ErrHeaderLengthUnexpected
		return
	}

	if extra == 0 {
		return
	}

	// Increase len of bts to extra bytes need to read.
	// Overwrite first 2 bytes that was read before.
	bts, err = io.ReadAll(io.LimitReader(r, int64(extra)))
	if err != nil {
		return
	}

	switch {
	case length == 126:
		h.Length = int64(binary.BigEndian.Uint16(bts[:2]))
		bts = bts[2:]

	case length == 127:
		if bts[0]&0x80 != 0 {
			err = ws.ErrHeaderLengthMSB
			return
		}
		h.Length = int64(binary.BigEndian.Uint64(bts[:8]))
		bts = bts[8:]
	}

	if h.Masked {
		copy(h.Mask[:], bts)
	}

	return
}
