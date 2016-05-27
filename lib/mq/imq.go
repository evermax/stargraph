package mq

// MessageQueue interface help masking the complexity of the amqp library
// and reduce them to the needs of this project. The interface helps testing
// and only provide with the ability to declare a queue, publish on a queue
// and consume the elements coming on a queue.
type MessageQueue interface {
	DeclareQueue(string) error
	Publish(string, []byte) error
	Consume(string, Receiver) error
}

// Receiver is a function type that will be called every time
// a message arrives on the message queue. It will be provided
// with a Delivery. The Receiver can then use the Body of the
// Delivery to make the necessary computations and either
// acknowledge or non acknowledge the message for it to be
// requeued or not.
// The boolean channel can make the message queue stop consuming the queue.
// If so, it will be needed to call Consume again.
type Receiver func(Delivery, chan bool)

// Delivery interface provide a wrapper for the message and acknowledgment
// system of AMQP. It will allows the read the body of the message and either
// aknowledge it or non acknowledge it.
type Delivery interface {
	Body() []byte
	Ack(bool) error
	Nack(bool, bool) error
}
