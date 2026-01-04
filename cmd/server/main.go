package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.1.20:5672/")
	FailOnError(err, "Failed to connect to AMQP")
	defer func(conn *amqp.Connection) {
		err = conn.Close()
		FailOnError(err, "Unable to close AMQP connection")
	}(conn)

	log.Println("Connection to AMQP ok")

	mq, err := conn.Channel()
	FailOnError(err, "Error opening channel")

	err = pubsub.PublishJSON(mq, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	fmt.Println("Starting Peril server...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Exiting")
}
