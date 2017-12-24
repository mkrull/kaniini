// Testing kaniini is mostly integration testing as it is a pretty thin layer
// above amqp and most tests involve connecting to RabbitMQ. To avoid that
// basically requires to write a mock of RabbitMQ.

package kaniini

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestQueueConstruction(t *testing.T) {
	consumer, err := NewQueue(
		"amqp://guest:guest@localhost:5672",
		"kaniini")
	consumer.Done() <- struct{}{}
	if err != nil {
		t.Errorf("Could not create consumer: %s", err)
	}
}

func TestMain(m *testing.M) {
	expected := "test message"

	consumer, _ := NewQueue(
		"amqp://guest:guest@localhost:5672",
		"test-integration")

	_ = consumer.Send([]byte("test message"))

	timer := time.NewTimer(time.Second * 1)

	select {
	case message := <-consumer.Receive():
		message.Ack()
		msg, _ := message.(*delivery)
		consumer.Done() <- struct{}{}
		fmt.Printf("Got: %s", string(msg.Body))
		if string(msg.Body) != expected {
			os.Exit(1)
		}
		os.Exit(m.Run())
	case <-timer.C:
		consumer.Done() <- struct{}{}
		fmt.Println("Timeout while consuming messages.")
		os.Exit(1)
	}
}
