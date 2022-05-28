package main

import (
	"flag"
	"github.com/gobwas/ws"
	"github.com/google/gops/agent"
	"log"
	_ "net/http/pprof"
	"ws/msg"
	"ws/wsgateway"
)

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatalf("agent.Listen err: %v", err)
	}

	flag.Parse()
	log.SetFlags(0)
	ch := make(chan int)
	s := wsgateway.Init(*addr)
	wsgateway.AddHandler("1", wsgateway.HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *wsgateway.Conn) {
		conn.WriteMsg(ws.OpText, []byte("hahaha"))
	}))
	s.Start(false)
	<-ch
}
