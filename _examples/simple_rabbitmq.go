package main

import (
	"fmt"

	"github.com/mkrull/kaniini"
)

func main() {
	queue, _ := kaniini.NewQueue(
		"amqp://guest:guest@localhost:5672",
		"kaniini")

	_ = queue.Send([]byte("test message"))
	select {
	case msg := <-queue.Receive():
		defer msg.Ack()
		fmt.Println(string(msg.Body))
	}
}
