package main

import (
	"flag"
	"fmt"
	"github.com/panjf2000/gnet"
	"log"
)

type wsCodec struct {
}

func (w wsCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (w wsCodec) Decode(c gnet.Conn) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

type echoServer struct {
	*gnet.EventServer
}

func (es *echoServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("Echo server is listening on %s (multi-cores: %t, loops: %d)\n",
		srv.Addr.String(), srv.Multicore, srv.NumLoops)
	return
}
func (es *echoServer) React(c gnet.Conn) (out []byte, action gnet.Action) {

	// Echo synchronously.
	c.AsyncWrite([]byte("haha"))
	return

	/*
		// Echo asynchronously.
		data := append([]byte{}, frame...)
		go func() {
			time.Sleep(time.Second)
			c.AsyncWrite(data)
		}()
		return
	*/
}

// OnOpened fires when a new connection has been opened.
// The info parameter has information about the connection such as
// it's local and remote address.
// Use the out return value to write data to the connection.
func (es *echoServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

// OnClosed fires when a connection has been closed.
// The err parameter is the last known connection error.
func (es *echoServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	return gnet.None
}

func main() {
	var port int
	var multicore bool

	// Example command: go run echo.go --port 9000 --multicore=true
	flag.IntVar(&port, "port", 9000, "--port 9000")
	flag.BoolVar(&multicore, "multicore", false, "--multicore true")
	flag.Parse()
	echo := new(echoServer)
	log.Fatal(gnet.Serve(echo, fmt.Sprintf("tcp://:%d", port), gnet.WithMulticore(multicore)))
}
