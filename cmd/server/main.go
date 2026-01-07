package server

import (
	"fmt"
	"log"

	"github.com/oleksii-kalinin/learn-pub-sub-starter/internal/gamelogic"
	"github.com/oleksii-kalinin/learn-pub-sub-starter/internal/pubsub"
	"github.com/oleksii-kalinin/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Serve(amqpUrl string) error {
	conn, err := amqp.Dial(amqpUrl)
	FailOnError(err, "Failed to connect to AMQP")
	defer func(conn *amqp.Connection) {
		err = conn.Close()
		if err != nil {
			log.Printf("Warning: Unable to close AMQP connection: %s", err)
		}
	}(conn)

	publishCh, err := conn.Channel()
	FailOnError(err, "Error opening channel")
	defer publishCh.Close()

	err = pubsub.Subscribe(conn, routing.ExchangePerilTopic, routing.GameLogSlug, "*", pubsub.DurableQueue, handlerLog, pubsub.GobUnmarshal[routing.GameLog])
	if err != nil {
		FailOnError(err, "unable to subscribe to logs")
	}

	fmt.Println("Starting Peril server...")

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) <= 0 {
			continue
		}
		switch input[0] {
		case "pause":
			log.Println("sending pause message")
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			FailOnError(err, "")

		case "resume":
			log.Println("sending resume message")
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			FailOnError(err, "")

		case "quit":
			log.Println("quitting")
			return err
		default:
			log.Println("unknown command")
			continue
		}
	}
}
