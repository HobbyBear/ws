package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/url"
	"nfw/nfw_base/utils/commfunc"
	"os"
	"syscall"
	"time"
	"ws"
)

func RedirectStderr() (err error) {
	logFile, err := os.OpenFile("./test-error.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	err = syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		return
	}
	return
}

func main() {

	RedirectStderr()
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/"}
	log.Printf("connecting to %s", u.String())
	for i := 1; i <= 20000; i++ {
		go func() {
			var (
				c   *websocket.Conn
				err error
			)
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
			//for i := 1; i <= 2000; i++ {
			//	p.Submit(func() {
			//		for {
			//			err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*2))
			//			if err != nil {
			//				log.Println(err, "客户端心跳")
			//
			//			}
			//			rand.Seed(time.Now().Unix())
			//			time.Sleep(time.Second * 3)
			//		}
			//	})
			//}
			//p.Submit(func() {
			//	for {
			//		mt, data, err := c.ReadMessage()
			//		if err != nil {
			//			if err.Error() != "websocket: close 1006 (abnormal closure): unexpected EOF" {
			//				log.Println(err, string(data), mt)
			//			}
			//			time.Sleep(2 * time.Second)
			//		}
			//		fmt.Println(string(data))
			//	}
			//})
			//p.Submit(func() {
			//	for {
			//		err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*2))
			//		if err != nil {
			//			log.Println(err, "客户端心跳")
			//
			//		}
			//		rand.Seed(time.Now().Unix())
			//		time.Sleep(time.Second * 3)
			//	}
			//})
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Println(err)
						return
					}
				}()

				for {

					data := ws.DataMsg{
						MsgType: "1",
						Content: []byte("haha"),
					}
					err := c.WriteMessage(websocket.TextMessage, data.MarshalJSON())
					if err != nil {
						log.Println("write:", err)
						return
					}
					time.Sleep(2 * time.Second)
				}
			}()
		}()

	}

	fmt.Println("连接client finished")
	time.Sleep(10 * time.Hour)
}
