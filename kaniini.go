package kaniini

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

type Queue interface {
	Receive() <-chan Delivery
	Done() chan struct{}
	Send([]byte) error
}

type Delivery interface {
	Unmarshal([]byte) error
	Ack() error
	Error() error
	SaveRaw(interface{}) error
}

type delivery struct {
	amqpDelivery amqp.Delivery
	Body         []byte
}

func (d *delivery) Ack() error {
	return d.amqpDelivery.Ack(true)
}

func (d *delivery) Error() error {
	return nil
}

func (d *delivery) Unmarshal(data []byte) error {
	d.Body = data

	return nil
}

func (d *delivery) SaveRaw(raw interface{}) error {
	amqpDelivery, ok := raw.(amqp.Delivery)

	if ok {
		d.amqpDelivery = amqpDelivery
		return nil
	} else {
		return errors.New("Raw message is no amqp.Delivery")
	}
}

type queue struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	Deliveries <-chan Delivery
	done       chan struct{}
	uri        string
	name       string
}

func mapFunc(amqpDelivery <-chan amqp.Delivery, deliveries chan Delivery) {
	for msg := range amqpDelivery {
		d := &delivery{}
		d.Unmarshal(msg.Body)
		d.SaveRaw(msg)
		deliveries <- d
	}
}

func mapDelivery(amqpDelivery <-chan amqp.Delivery, mapFunc func(<-chan amqp.Delivery, chan Delivery)) chan Delivery {
	deliveries := make(chan Delivery)
	go mapFunc(amqpDelivery, deliveries)

	return deliveries
}

func NewQueue(uri string, name string) (Queue, error) {
	consumer := &queue{
		name: name,
		uri:  uri,
		done: make(chan struct{}),
	}

	err := consumer.declare()
	if err != nil {
		return nil, err
	}

	// Start consuming
	del, err := consumer.channel.Consume(
		consumer.name,
		"consumer_tag",
		false,
		false,
		false,
		false,
		nil)

	if err != nil {
		return nil, err
	}

	consumer.Deliveries = mapDelivery(del, mapFunc)

	// Close when done
	consumer.closeRoutine()

	return consumer, nil
}

func (q *queue) closeRoutine() {
	go func() {
		for {
			select {
			case <-q.done:
				fmt.Println("Closing connections")
				q.stop()
				q.done <- struct{}{}
				return
			}
		}
	}()
}

func (q *queue) stop() error {
	err := q.channel.Close()
	if err != nil {
		return err
	}

	return q.conn.Close()
}

func (q *queue) declare() error {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(q.uri)
	if err != nil {
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		return err
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		q.name,
		"direct",
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	// Declare and bind queue
	que, err := channel.QueueDeclare(
		q.name,
		true,
		false,
		false,
		false,
		nil)

	err = channel.QueueBind(
		que.Name,
		"",
		q.name,
		false,
		nil)
	if err != nil {
		return err
	}

	q.conn = conn
	q.channel = channel

	return nil
}

func (q *queue) Receive() <-chan Delivery {
	return q.Deliveries
}

func (q *queue) Done() chan struct{} {
	return q.done
}

func (q *queue) Send(msg []byte) error {
	return q.channel.Publish(q.name, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        msg,
	})
}
