package main

import (
	"fmt"
	"io"
	"log"
	"strings"
)

func main() {
	r := strings.NewReader("abcdef")
	buf := make([]byte, 2, 100)

	for i := 1; i <= 2; i++ {
		n, err := io.ReadFull(r, buf)
		if err != nil {
			log.Println(err, string(buf), n)
		} else {
			fmt.Println(n, string(buf))
		}
	}
}
