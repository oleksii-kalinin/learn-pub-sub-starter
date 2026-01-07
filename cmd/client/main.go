package client

import (
	"fmt"
	"log"
	"strconv"

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

func Start(amqpUrl string) error {
	client, err := amqp.Dial(amqpUrl)
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
		return err
	}
	defer publishCh.Close()
	fmt.Println("Starting Peril client...")

	msg, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Println(err)
	}
	state := gamelogic.NewGameState(msg)

	err = pubsub.Subscribe(client, routing.ExchangePerilDirect, routing.PauseKey+"."+state.GetUsername(), routing.PauseKey, pubsub.TransientQueue, handlerPause(state, publishCh), pubsub.JsonUnmarshal[routing.PlayingState])
	if err != nil {
		FailOnError(err, "unable to subscribe to pause")
	}

	err = pubsub.Subscribe(client, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+state.GetUsername(), routing.ArmyMovesPrefix+".*", pubsub.TransientQueue, handlerMove(state, publishCh), pubsub.JsonUnmarshal[gamelogic.ArmyMove])
	if err != nil {
		FailOnError(err, "unable to subscribe to move")
	}

	err = pubsub.Subscribe(client, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+"."+state.GetUsername(), pubsub.DurableQueue, handlerWar(state, publishCh), pubsub.JsonUnmarshal[gamelogic.RecognitionOfWar])
	if err != nil {
		FailOnError(err, "unable to subscribe to war")
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
			if len(words) < 2 {
				log.Println("spam count should be provided")
				continue
			}
			cnt, err := strconv.ParseInt(words[1], 10, 32)
			if err != nil {
				log.Println("unable to parse count")
				continue
			}
			if cnt <= 0 {
				log.Println("count must be positive")
				continue
			}
			for _ = range cnt {
				msg := gamelogic.GetMaliciousLog()
				err = publishGameLog(publishCh, state, msg)
				if err != nil {
					log.Println("unable to send spam message")
					continue
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			return err
		default:
			log.Println("unknown command")
		}
	}
}
