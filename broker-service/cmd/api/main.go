package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {

	rabbitCon, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitCon.Close()

	app := Config{
		Rabbit: rabbitCon,
	}

	log.Printf("Starting Broker Service on port %s \n", webPort)

	//define http Server
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	//start the http server
	err = svr.ListenAndServe()
	if err != nil {
		log.Panicf("Error happened while connecting to server.. Details : %s \n", err)
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
