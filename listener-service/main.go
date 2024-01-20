package main

import (
	"listener-service/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	rabbitCon, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitCon.Close()

	consumer, err := event.NewConsumer(rabbitCon)

	if err != nil {
		log.Println("Error while creating new consumer")
		log.Println(err)
		os.Exit(1)
	}
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println("Error while listening")
		log.Println(err)
		os.Exit(1)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("RabbitMQ is not ready to connect .. ")
			counts++
		} else {
			connection = c
			break
		}

		if counts > 5 {
			log.Println("Number of RabbitMQ connection trials exceeded")
			log.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing Off...")
		time.Sleep(backOff)
		continue
	}

	log.Println("RabbitMQ Connection is established ")
	return connection, nil
}
