package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/url"
	"nfw/nfw_base/utils/commfunc"
	"time"
	"ws"
)

func main() {

	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/"}
	log.Printf("connecting to %s", u.String())
	for i := 1; i <= 10000; i++ {
		go func() {
			var (
				c   *websocket.Conn
				err error
			)
			time.Sleep(30 * time.Millisecond)
			commfunc.Retry(func() bool {
				c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					log.Println("dial:", err, i, u.String())
					return false
				}
				return true
			}, nil, 3, 1)
			if err != nil {
				return
			}
			time.Sleep(time.Duration(rand.Int31n(10)) * time.Second)
			go func() {
				for {
					mt, data, err := c.ReadMessage()
					if err != nil {
						if err.Error() != "websocket: close 1006 (abnormal closure): unexpected EOF" {
							log.Println(err, string(data), mt)
						}
						time.Sleep(2 * time.Second)
					}
				}
			}()
			go func() {
				for {
					err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*2))
					if err != nil {
						log.Println(err, "客户端心跳")

					}
					rand.Seed(time.Now().Unix())
					time.Sleep(time.Minute * 3)
				}
			}()
			go func() {
				timer := time.NewTimer(5 * time.Second)
				timer.Reset(time.Duration(rand.Int31n(10)) * time.Second)
				for {
					select {
					case <-timer.C:
						data := ws.DataMsg{
							MsgType: "1",
							Content: []byte("haha"),
						}
						err := c.WriteMessage(websocket.TextMessage, data.MarshalJSON())
						if err != nil {
							log.Println("write:", err)
							return
						}
						timer.Reset(time.Duration(rand.Int31n(10)) * time.Second)
					}
				}
			}()

		}()
	}
	fmt.Println("连接client finished")
	time.Sleep(10 * time.Hour)
}
