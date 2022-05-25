package main

import (
	"flag"
	ws2 "github.com/gobwas/ws"
	"github.com/google/gops/agent"
	"log"
	_ "net/http/pprof"
	"ws"
)

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatalf("agent.Listen err: %v", err)
	}

	flag.Parse()
	log.SetFlags(0)
	ch := make(chan int)
	s := ws.InitWs(*addr)
	ws.GetRouterMgr().RegHandler("1", func(req *ws.RouterHandlerReq) {
		err := req.Conn.WriteMsg(&ws.RawMsg{WsMsgType: ws2.OpText, Content: []byte(req.Content)})
		if err != nil {
			log.Println(err)
		}
	})
	s.Start(false)
	<-ch
}
