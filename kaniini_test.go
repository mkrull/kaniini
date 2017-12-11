// Testing kaniini is mostly integration testing as it is a pretty thin layer
// above amqp and most tests involve connecting to RabbitMQ. To avoid that
// basically requires to write a mock of RabbitMQ.

// +build integration

package kaniini

import (
	"testing"
	"time"
)

func TestConsumerConstruction(t *testing.T) {
	consumer, err := NewConsumer(
		"amqp://guest:guest@localhost:5672",
		"kaniini")
	defer consumer.Done()

	if err != nil {
		t.Errorf("Could not create consumer.")
	}
}

func TestConsumer(t *testing.T) {
	expected := "test message"

	consumer, err := NewConsumer(
		"amqp://guest:guest@localhost:5672",
		"kaniini")
	defer consumer.Done()
	if err != nil {
		t.Errorf("Could not create consumer: %s", err)
	}

	err = consumer.Emit([]byte("test message"))
	if err != nil {
		t.Errorf("Error emitting message: %s", err)
	}

	timer := time.NewTimer(time.Second * 1)

	select {
	case m := <-consumer.Chan():
		defer m.Ack()
		if string(m.Body) != expected {
			t.Errorf("Got %s, expected: %s", string(m.Body), expected)
		}
	case <-timer.C:
		t.Errorf("Timeout while consuming messages.")
	}
}
