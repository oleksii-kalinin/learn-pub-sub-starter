package main

import (
	"fmt"
	"log"
	"os"

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

	state := gamelogic.NewGameState(msg)
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = state.CommandSpawn(words)
			if err != nil {
				log.Println(err)
			}
		case "move":
			_, err = state.CommandMove(words)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("move ok")
		case "status":
			state.CommandStatus()
		case "spam":
			log.Println("no spam!")
		case "quit":
			gamelogic.PrintQuit()
			//break StateLoop
			return
		default:
			log.Println("unknown command")
		}
	}
}
