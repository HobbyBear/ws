package main

import (
	"flag"
	"log"
	"ws"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	ch := make(chan int)
	s := ws.InitWs(*addr)
	ws.GetRouterMgr().RegHandler("haha", func(req *ws.RouterHandlerReq) {

	})
	s.Start()
	<-ch
}
