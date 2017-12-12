package kaniini

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryConstruction(t *testing.T) {
	assert := assert.New(t)
	q, ok := NewInMemoryQueue("kaniini", 10).(*inMemoryQueue)

	assert.True(ok, "Object is an in memory queue")
	assert.Equal(10, cap(q.deliveries), "Buffer should be of correct size")
}

func TestInMemorySending(t *testing.T) {
	assert := assert.New(t)
	q, _ := NewInMemoryQueue("kaniini", 10).(*inMemoryQueue)

	q.Send([]byte("test message"))
	assert.Equal(1, len(q.deliveries), "Queue contains one element")
}

func TestInMemoryReceiving(t *testing.T) {
	assert := assert.New(t)
	q, _ := NewInMemoryQueue("kaniini", 10).(*inMemoryQueue)
	expected := []byte("test message")

	q.Send(expected)

	timer := time.NewTimer(time.Second * 1)

	select {
	case msg := <-q.Receive():
		assert.Equal(expected, msg.Body, "Receiving expected message")
	case <-timer.C:
		t.Errorf("Timeout while consuming messages.")
	}
}
