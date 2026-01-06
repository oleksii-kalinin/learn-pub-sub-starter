package main

import (
	"fmt"
	"log"
	"os"

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

func main() {
	client, err := amqp.Dial(os.Getenv("AMQP_URL"))
	FailOnError(err, "Failed to connect to AMQP")
	defer func(client *amqp.Connection) {
		err = client.Close()
		if err != nil {
			log.Printf("Warning: Unable to close AMQP connection: %s", err)
		}
	}(client)
	publishCh, err := client.Channel()
	if err != nil {
		log.Println(err)
		return
	}
	defer publishCh.Close()
	fmt.Println("Starting Peril client...")

	msg, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Println(err)
	}
	state := gamelogic.NewGameState(msg)

	err = pubsub.SubscribeJSON(client, routing.ExchangePerilDirect, routing.PauseKey+"."+state.GetUsername(), routing.PauseKey, pubsub.TransientQueue, handlerPause(state))
	if err != nil {
		FailOnError(err, "unable to subscribe to pause")
	}

	err = pubsub.SubscribeJSON(client, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+state.GetUsername(), routing.ArmyMovesPrefix+".*", pubsub.TransientQueue, handlerMove(state))
	if err != nil {
		FailOnError(err, "unable to subscribe to move")
	}

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
			move, err := state.CommandMove(words)
			if err != nil {
				log.Println(err)
				continue
			}

			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+state.GetUsername(), move)
			if err != nil {
				log.Println(err)
				continue
			}
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
