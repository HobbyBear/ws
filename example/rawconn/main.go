package main

import (
	"io"
	"log"
	"net"

	"github.com/gobwas/ws"
)

func main() {
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		_, err = ws.Upgrade(conn)
		if err != nil {
			// handle error
			log.Fatal(err)
		}
		go func() {
			defer conn.Close()

			for {
				header, err := ws.ReadHeader(conn)
				if err != nil {
					// handle error
					log.Fatal(err)
				}

				payload := make([]byte, header.Length)
				_, err = io.ReadFull(conn, payload)
				if err != nil {
					// handle error
					log.Fatal(err)
				}
				if header.Masked {
					ws.Cipher(payload, header.Mask, 0)
				}
				// Reset the Masked flag, server frames must not be masked as
				// RFC6455 says.
				header.Masked = false

				if err := ws.WriteHeader(conn, header); err != nil {
					// handle error
					log.Fatal(err)
				}
				if _, err := conn.Write(payload); err != nil {
					// handle error
					log.Fatal(err)
				}

				if header.OpCode == ws.OpClose {
					return
				}
			}
		}()
	}
}
