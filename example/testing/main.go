package main

import "fmt"

func main() {
	ch := make(chan int, 10)
	ch <- 12
	ch <- 13
	close(ch)
	for c := range ch {
		fmt.Println(c)
	}
}
