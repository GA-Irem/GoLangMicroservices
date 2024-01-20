package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func (e *Emitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return declareExchange(channel)
}

func (emitter *Emitter) Push(event string, severity string) error {

	channel, err := emitter.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	log.Println("Pushing message to RabbitMQ Channel")

	err = channel.Publish(
		"special_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
	if err != nil {
		log.Println("Error while publishing to RabbitMQ Channel")
		log.Println(err)
		return err
	}

	log.Println("Message Published to RabbitMQ Channel")
	return nil
}

func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	log.Println("NewEmitter is executed..")
	emitter := Emitter{
		connection: conn,
	}

	err := emitter.setup()
	if err != nil {
		log.Println("Error While registering emitter..")
		log.Println(err)
		return Emitter{}, err
	}
	return emitter, nil
}
