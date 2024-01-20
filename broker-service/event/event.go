package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// declareExchange declares a new exchange on the given AMQP channel.
//
// Parameters:
// - ch: The AMQP channel on which to declare the exchange.
//
// Returns:
// - error: An error if the exchange declaration fails.
func declareExchange(ch *amqp.Channel) error {

	return ch.ExchangeDeclare(
		"special_topic", // name
		"topic",         //type
		true,            //durable?
		false,           //autodeleted
		false,           //internal usage only ?
		false,           //no-wait
		nil,             //agments){

	)
}

// declareRandomQueue declares a random queue.
//
// The function takes a pointer to an amqp.Channel as its parameter.
// It returns an amqp.Queue and an error.
func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"special_queue", // name
		false,           // durable
		false,           // delete when unused
		true,            // exclusive
		false,           // no-wait
		nil,             // arguments
	)
}
