package main

import (
	"flag"
	"github.com/google/gops/agent"
	"log"
	"time"
	"ws"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatalf("agent.Listen err: %v", err)
	}

	flag.Parse()
	log.SetFlags(0)
	ch := make(chan int)
	s := ws.InitWs(*addr)
	ws.GetRouterMgr().RegHandler("1", func(req *ws.RouterHandlerReq) {

	})
	s.Start()

	time.Sleep(10 * time.Second)
	s.ShutDown()
	<-ch
}
