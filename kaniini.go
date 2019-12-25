package kaniini

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Exchange int

const (
	None  Exchange = iota
	Direct
	Topic
	Fanout
	Headers
)

func (e Exchange) String() string {
	return []string{
		"",
		"direct",
		"topic",
		"fanout",
		"headers",
	}[e]
}

type Queue interface {
	Receive() <-chan *Delivery
	Done() chan struct{}
	Send([]byte) error
}

type Delivery struct {
	amqpDelivery amqp.Delivery
	Body         []byte
}

func (d *Delivery) Ack() {
	d.amqpDelivery.Ack(true)
}

type queue struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	Deliveries   <-chan *Delivery
	done         chan struct{}
	uri          string
	name         string
	exchangeName string
	exchangeType Exchange
}

func mapDelivery(amqpDelivery <-chan amqp.Delivery) chan *Delivery {
	deliveries := make(chan *Delivery)
	go func() {
		for {
			select {
			case msg := <-amqpDelivery:
				deliveries <- &Delivery{
					Body:         msg.Body,
					amqpDelivery: msg,
				}
			}
		}
	}()

	return deliveries
}

func NewQueue(uri string, name string) (Queue, error) {
	q := &queue{
		name: name,
		uri:  uri,
		done: make(chan struct{}),
	}

	return newQueue(q)
}

func NewQueueOnExchange(uri string, name string, exchange string, exchangeT Exchange) (Queue, error) {
	q := &queue{
		name: name,
		uri:  uri,
		done: make(chan struct{}),
		exchangeName: exchange,
		exchangeType: exchangeT,
	}

	return newQueue(q)
}

func newQueue(q *queue) (Queue, error) {
	// set default exchange config
	if q.exchangeName == "" {
		q.exchangeName = q.name
	}

	if q.exchangeType == None {
		q.exchangeType = Direct
	}

	err := q.declare()
	if err != nil {
		return nil, err
	}

	// Start consuming
	del, err := q.channel.Consume(
		q.name,
		"consumer_tag",
		false,
		false,
		false,
		false,
		nil)

	if err != nil {
		return nil, err
	}

	q.Deliveries = mapDelivery(del)

	// Close when done
	q.closeRoutine()

	return q, nil
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
		q.exchangeName,
		q.exchangeType.String(),
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
		q.exchangeName,
		false,
		nil)
	if err != nil {
		return err
	}

	q.conn = conn
	q.channel = channel

	return nil
}

func (q *queue) Receive() <-chan *Delivery {
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
