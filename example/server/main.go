package main

import (
	"flag"
	"github.com/go-redis/redis"
	"github.com/gobwas/ws"
	"github.com/google/gops/agent"
	"log"
	_ "net/http/pprof"
	"ws/broker"
	"ws/msg"
	"ws/wsapi"
	"ws/wsgateway"
)

var addr = flag.String("addr", "0.0.0.0:8081", "http service address")

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatalf("agent.Listen err: %v", err)
	}

	flag.Parse()
	log.SetFlags(0)
	ch := make(chan int)
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	s := wsgateway.Init(*addr, wsgateway.OptionSetBroker(broker.NewRedisBroker(client)))
	wsgateway.AddHandler("1", wsgateway.HandlerFunc(func(opCode ws.OpCode, reqMsg *msg.ReqMsg, conn *wsgateway.Conn) {
		(&wsapi.Client{Producer: broker.NewRedisBroker(client)}).SendMsgToAll([]byte("无敌"), ws.OpText)
	}))
	s.Start(false)
	<-ch
}
