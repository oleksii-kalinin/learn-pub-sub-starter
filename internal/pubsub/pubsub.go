package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(val)
	if err != nil {
		log.Println(err)
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buffer.Bytes(),
	})
	return err
}

func Subscribe[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType, unmarshaller func([]byte) (T, error)) error {
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
			v, err := unmarshaller(msg.Body)
			if err != nil {
				log.Printf("unmarshal error: %v", err)
				err = msg.Nack(false, false)
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
				log.Printf("ack/nack failed: %v", err)
			}
		}
	}()

	return nil
}

func JsonUnmarshal[T any](data []byte) (T, error) {
	var target T
	err := json.Unmarshal(data, &target)
	return target, err
}

func GobUnmarshal[T any](data []byte) (T, error) {
	var target T
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(&target)
	return target, err
}
