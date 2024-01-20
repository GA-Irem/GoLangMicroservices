package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {

	consumer := Consumer{
		conn: conn,
	}
	err := consumer.register()
	if err != nil {
		return Consumer{}, err
	}
	return consumer, nil
}

func (consumer *Consumer) register() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		log.Println("Error while registering Consumer to Rabbit MQ Channel")
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Data string `json:"data"`
	Name string `json:"name"`
}

func (consumer *Consumer) Listen(topics []string) error {

	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := declareRandomQueue(ch)

	if err != nil {
		log.Println("Error while declaring Random Queue")
		log.Println(err)
		return err
	}

	for _, s := range topics {
		ch.QueueBind(
			q.Name,          // queue name
			s,               // routing key
			"special_topic", // exchange
			false,
			nil,
		)

		if err != nil {
			log.Println("Error while binding Queue")
			log.Println(err)
			return err
		}
	}

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for exchange Message [Exchange, Queue] [special_topic, %s] \n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "eventLog":
		log.Println("Do Sth with Log in broker-service")
		err := logEvent(payload)
		if err != nil {
			log.Println("Error while logging Event")
			log.Println(err)
		}
	case "auth":
		log.Println("Do Sth with Auth")
	case "mail":
		log.Println("Do Sth with Mail")
	default:
		log.Println("Unknown service")
		log.Println(payload.Name)
	}
}

func logEvent(entry Payload) error {

	//only in development, production : use Indent
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))

	if err != nil {
		log.Println("Error while logging Event")
		log.Println(err)
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		log.Println("Error After calling log service")
		log.Println(err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Println("Response in not in expected status")
		log.Println(err)
		return err
	}

	return nil
}
