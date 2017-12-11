package kaniini

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Consumer interface {
	Chan() <-chan *Delivery
	Done() chan struct{}
	Emit([]byte) error
}

type Delivery struct {
	amqpDelivery amqp.Delivery
	Body         []byte
}

func (d *Delivery) Ack() {
	d.amqpDelivery.Ack(true)
}

type consumer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	Deliveries <-chan *Delivery
	done       chan struct{}
	uri        string
	name       string
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

func NewConsumer(uri string, name string) (Consumer, error) {
	consumer := &consumer{
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

	consumer.Deliveries = mapDelivery(del)

	// Close when done
	consumer.closeRoutine()

	return consumer, nil
}

func (c *consumer) closeRoutine() {
	go func() {
		for {
			select {
			case <-c.done:
				fmt.Println("Closing connections")
				c.stop()
				c.done <- struct{}{}
				return
			}
		}
	}()
}

func (c *consumer) stop() error {
	err := c.channel.Close()
	if err != nil {
		return err
	}

	return c.conn.Close()
}

func (c *consumer) declare() error {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(c.uri)
	if err != nil {
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		return err
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		c.name,
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
	queue, err := channel.QueueDeclare(
		c.name,
		true,
		false,
		false,
		false,
		nil)

	err = channel.QueueBind(
		queue.Name,
		"",
		c.name,
		false,
		nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.channel = channel

	return nil
}

func (c *consumer) Chan() <-chan *Delivery {
	return c.Deliveries
}

func (c *consumer) Done() chan struct{} {
	return c.done
}

func (c *consumer) Emit(msg []byte) error {
	return c.channel.Publish(c.name, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        msg,
	})
}
