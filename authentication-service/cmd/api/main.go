package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var count int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {

	log.Println("Starting Authentication Services")

	//TODO connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Failed while connecting to Postgress !!")
	}

	//Set Config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}
	log.Printf("Starting Authentication Service on port %s \n", webPort)

	//define http Server
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	//start the http server
	err := svr.ListenAndServe()
	if err != nil {
		log.Panicf("Error happened while connecting to Authentication server.. Details : %s \n", err)
	}
}

func openDBConn(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	for {
		connection, err := openDBConn(dsn)
		if err != nil {
			log.Println("DB Connection is initializing!")
			count++
		} else {
			log.Println("DB connection is established to postgress")
			return connection
		}

		if count > 20 {
			log.Println("Timeout happened while establishing the connection")
			log.Println(err)
			return nil
		}

		log.Println("Sleeping for 2 more seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}
