package main

import (
	"fmt"

	"github.com/mkrull/kaniini"
)

func main() {
	consumer, _ := kaniini.NewConsumer(
		"amqp://guest:guest@localhost:5672",
		"kaniini")

	_ = consumer.Emit([]byte("test message"))
	select {
	case msg := <-consumer.Chan():
		defer msg.Ack()
		fmt.Println(string(msg.Body))
	}
}
