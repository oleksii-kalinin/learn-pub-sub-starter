package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	DurableQueue   SimpleQueueType = "durable"
	TransientQueue SimpleQueueType = "transient"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		log.Panicf("Unable to marshal to json: %s", err)
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonVal,
	})
	return err
}

// SubscribeJSON subscribes to the exchange with queueName and key for the routing
func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		log.Println("unable to bind")
		return err
	}

	consumer, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Println("unable to create consumer")
		return err
	}
	go func() {
		defer channel.Close()
		for msg := range consumer {
			var v T

			err = json.Unmarshal(msg.Body, &v)
			if err != nil {
				log.Println("unable to parse json")
				continue
			}
			ackType := handler(v)
			switch ackType {
			case Ack:
				log.Println("calling ack")
				err = msg.Ack(false)
			case NackRequeue:
				log.Println("calling Neg ack requeue")
				err = msg.Nack(false, true)
			case NackDiscard:
				log.Println("calling Neg ack discard")
				err = msg.Nack(false, false)
			default:
				log.Println("wrong ack event")
			}

			if err != nil {
				log.Println("unable to parse json")
			}
		}
	}()

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := queueType == DurableQueue
	autoDelete := queueType == TransientQueue
	exclusive := queueType == TransientQueue

	qTable := amqp.Table{}

	qTable["x-dead-letter-exchange"] = "peril_dlx"

	q, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, false, qTable)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(q.Name, key, exchange, false, nil)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}
