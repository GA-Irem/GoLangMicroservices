package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort = "80"
	rpcPort = "5001"
	//local mongoURL = "mongodb://localhost:27017"
	// to try the connection on local
	// go to logger service logger-service % go run ./cmd/api
	//deployment
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {

	mongoClient, err := connectToMongo()

	if err != nil {
		log.Panic(err)
		return
	}

	client = mongoClient

	// create a context to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	//close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	//Register RPC Server
	err = rpc.Register(new(RPCServer))
	log.Println("listenRPC is called")
	go app.listenRPC()

	//register gRPC
	go app.gRPCListen()

	//start web server

	log.Println("Starting Mongo on Port ", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) listenRPC() error {
	log.Println("Starting RPC server on port ", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))

	if err != nil {
		log.Panic("Failed to listen on port ", rpcPort)
		log.Println(err)
		return err
	}
	defer listen.Close()

	for {
		log.Println("Trying to accept RPC connection", rpcPort)
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		log.Println("Accepting RPC connection", rpcPort)
		go rpc.ServeConn(rpcConn)
	}
}

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}

	log.Println("Connected to mongo!")

	return c, nil
}
