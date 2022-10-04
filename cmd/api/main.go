package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"
var counts int64

// type Config struct{
// 	Repo data.Repository
// 	Rabbit *amqp.Connection
// }

type Config struct{
	Repo data.Repository
	Client *http.Client
}

func main() {
	// connect to rabbitMQ
	// rabbitConn,err := connectToRabbitMQ()
	// if err != nil {
	// 	log.Println(err)
	// 	os.Exit(1)
	// }

	// defer rabbitConn.Close()
	// log.Println("connected to RabbitMQ")

	//connect to db
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to DB")
	}

	//set up config
	// app := Config{
	// 	Rabbit: rabbitConn,
	// }

	app := Config{
		Client: &http.Client{},
	}

	//define http server
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Printf("Starting authentication service at port %s", webPort)

	//start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil,err
	}
	
	err = db.Ping()
	if err != nil {
		return nil,err
	}

	return db,nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection , err := openDB(dsn)

		if err != nil {
			log.Println("Postgres not yet ready...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing of for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	var counts int64
	var backoff = 1 * time.Second
	var connection *amqp.Connection

	// dont continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil,err
		}

		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off..")
		time.Sleep(backoff)
		continue
	}

	return connection, nil
}

func (app *Config) SetupRepo(conn *sql.DB) {
	db := data.NewPostgresRepository(conn)
	app.Repo = db
}