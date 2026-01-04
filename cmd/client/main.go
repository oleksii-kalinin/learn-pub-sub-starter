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
	client, err := amqp.Dial(os.Getenv("AMQP_URL"))
	FailOnError(err, "Failed to connect to AMQP")
	defer func(client *amqp.Connection) {
		err = client.Close()
		FailOnError(err, "Unable to close AMQP connection")
	}(client)
	fmt.Println("Starting Peril client...")

	msg, err := gamelogic.ClientWelcome()
	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, msg)

	ch, _, err := pubsub.DeclareAndBind(client, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.TransientQueue)
	defer ch.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Exiting")
}
