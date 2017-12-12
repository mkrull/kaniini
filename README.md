# Kaniini

[![Build Status](https://travis-ci.org/mkrull/kaniini.svg?branch=master)](https://travis-ci.org/mkrull/kaniini)
[![codecov](https://codecov.io/gh/mkrull/kaniini/branch/master/graph/badge.svg)](https://codecov.io/gh/mkrull/kaniini)

Kaniini is a thin wrapper around [streadway ampq](github.com/streadway/amqp) for
easy use cases not involving routing logic.

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

## TODO:

- [ ] reconnect
