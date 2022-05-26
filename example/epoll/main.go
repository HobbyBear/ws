package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	r := strings.NewReader("haha")
	rr := bufio.NewReaderSize(r, 2)
	//buf, err := rr.Peek(1)
	//fmt.Println(buf, err)
	//buf, err = rr.Peek(1)
	//fmt.Println(buf, err)
	buf := make([]byte, 4)
	b, err := rr.Read(buf)
	fmt.Println(buf, err, b)
}
