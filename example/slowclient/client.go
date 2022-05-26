package main

import (
	"context"
	"github.com/gobwas/ws"
	"log"
)

func main() {
	conn, _, _, err := ws.Dial(context.TODO(), "ws://127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	f := ws.NewFrame(ws.OpContinuation, false, []byte("测试"))
	ws.WriteHeader(conn, f.Header)
	//time.Sleep(10 * time.Second)
	conn.Write(f.Payload)
	f = ws.NewFrame(ws.OpText, true, []byte("测试222"))
	ws.WriteHeader(conn, f.Header)
	conn.Write(f.Payload)
}
