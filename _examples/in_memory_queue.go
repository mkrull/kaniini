package main

import (
	"fmt"

	"github.com/mkrull/kaniini"
)

func main() {
	queue := kaniini.NewInMemoryQueue("kaniini", 10)

	queue.Send([]byte("test message"))
	select {
	case msg := <-queue.Receive():
		fmt.Println(string(msg.Body))
	}
}
