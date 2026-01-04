package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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
	server, err := amqp.Dial(os.Getenv("AMQP_URL"))
	FailOnError(err, "Failed to connect to AMQP")
	defer func(conn *amqp.Connection) {
		err = conn.Close()
		FailOnError(err, "Unable to close AMQP connection")
	}(server)

	mq, err := server.Channel()
	FailOnError(err, "Error opening channel")
	fmt.Println("Starting Peril server...")

	gamelogic.PrintServerHelp()
GameLoop:
	for {
		input := gamelogic.GetInput()
		if len(input) <= 0 {
			continue
		}
		switch input[0] {
		case "pause":
			log.Println("sending pause message")
			err = pubsub.PublishJSON(mq, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			FailOnError(err, "")

		case "resume":
			log.Println("sending resume message")
			err = pubsub.PublishJSON(mq, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			FailOnError(err, "")

		case "quit":
			log.Println("quitting")
			break GameLoop
		default:
			log.Println("unknow command")
			break
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Exiting")
}
