# Kaniini

Kaniini is a thin wrapper around [streadway ampq](github.com/streadway/amqp) for
easy use cases not involving routing logic.

    package main

    import (
        "fmt"

        "github.com/mkrull/kaniini"
    )

    func main() {
        consumer, _ := kaniini.NewConsumer(
            "amqp://guest:guest@localhost:5672",
            "kaniini")

        for {
            select {
            case msg := <-consumer.Chan():
                fmt.Println(string(msg.Body))
            }
        }
    }

TODO:
[] reconnect
