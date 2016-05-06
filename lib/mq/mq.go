package mq

import (
	"fmt"
	"github.com/evermax/stargraph/lib/store"
	"github.com/evermax/stargraph/service"
	"github.com/streadway/amqp"
)

type MQ struct {
	Conn        *amqp.Connection
	Channel     *amqp.Channel
	CreateQueue string
	UpdateQueue string
}

// connect to AMQP server, Channel
func NewMQ(amqpURL, addQueueN, updateQueueN string) (mq MQ, err error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return
	}
	ch, err := conn.Channel()
	if err != nil {
		return
	}
	addq, err := ch.QueueDeclare(
		addQueueN, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return
	}
	upq, err := ch.QueueDeclare(
		updateQueueN, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return
	}
	mq = MQ{
		Conn:        conn,
		Channel:     ch,
		CreateQueue: addq.Name,
		UpdateQueue: upq.Name,
	}
	return
}

// DeclareQueue declare a queue an set QoS
func (mq MQ) DeclareQueue(queueName string) error {
	_, err := mq.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %v", err)
	}

	return mq.Channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
}

func (mq MQ) Publish(queueName string, body []byte) error {
	return mq.Channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
}

func (mq MQ) Consume(stre store.Store, jobQueue chan service.Job, queueName string, r Receiptor) error {
	msgs, err := mq.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			msg := Message{
				delivery: d,
			}
			r(stre, jobQueue, msg, forever)
		}
		forever <- true
	}()
	<-forever
	return nil
}

type Receiptor func(store.Store, chan service.Job, Delivery, chan bool)

// Delivery interface provide
type Delivery interface {
	Body() []byte
	Ack(bool) error
	Nack(bool, bool) error
}

// Message is the wrapper for the Delivery struct of the github.com/streadway/amqp library.
// Its purpose is to met the be compliant with the Delivery interface and help making the
// rest of the project testable.
type Message struct {
	delivery amqp.Delivery
}

// Body will return the body of the message.
func (m Message) Body() []byte {
	return m.delivery.Body
}

// Ack delivers an acknowledgment that the message has been receive and treated.
// The multiple argument is true when the all the previous messages can be acknowledged as well.
func (m Message) Ack(multiple bool) error {
	return nil
}

// Nack delivers a negative acknowledgement signifying a failure in treating the message.
// If multiple is true, all the previous messages that weren't aknowledged yet are going
// to be negatively aknowledged.
// If requeue is true, it means that the message needs to be requeued.
func (m Message) Nack(multiple, requeue bool) error {
	return nil
}
