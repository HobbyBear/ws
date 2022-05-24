package main

import (
	"flag"
	"fmt"
	"github.com/google/gops/agent"
	"github.com/gorilla/websocket"
	"log"
	_ "net/http/pprof"
	"time"
	"ws"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatalf("agent.Listen err: %v", err)
	}

	flag.Parse()
	log.SetFlags(0)
	ch := make(chan int)
	s := ws.InitWs(*addr)
	ws.GetRouterMgr().RegHandler("1", func(req *ws.RouterHandlerReq) {
		err := req.Conn.WriteMsg(&ws.RawMsg{WsMsgType: websocket.TextMessage, Content: []byte("haha")})
		if err != nil {
			log.Println(err)
		}
	})
	s.Start(false)

	time.Sleep(30 * time.Second)
	start := time.Now()
	ws.SendMsgToAll([]byte("我读无所所所所所所"))
	fmt.Println("推送耗时", time.Now().Sub(start))
	<-ch
}
