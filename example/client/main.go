package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"ws"
)

func main() {

	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	log.Printf("connecting to %s", u.String())

	for i := 1; i <= 10000; i++ {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}

		go func() {
			mt, data, err := c.ReadMessage()
			if err != nil {
				log.Println(err, string(data), mt)
			}
		}()

		for j := 1; j <= 1000; j++ {
			data := ws.DataMsg{
				MsgId:   "1",
				Content: []byte("haha"),
			}
			err := c.WriteMessage(websocket.TextMessage, data.MarshalJSON())
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}

}
