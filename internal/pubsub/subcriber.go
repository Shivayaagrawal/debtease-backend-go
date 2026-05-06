package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack         AckType = iota // Simple Ack
	NackRequeue                // Nack and requeue
	NackDiscard                // Nack and discard
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Qos(10, 0, false); err != nil { // process one message at a time
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	_, _, err = DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("failed to declare and bind: %w", err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		defer ch.Close()
		for delivery := range msgs {
			msg, err := unmarshaller(delivery.Body)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				delivery.Nack(false, true)
				continue
			}

			ackType := handler(msg)
			switch ackType {
			case Ack:
				if err := delivery.Ack(false); err != nil {
					log.Printf("Failed to Ack message: %v", err)
				} else {
					log.Println("Message acknowledged (Ack)")
				}
			case NackRequeue:
				if err := delivery.Nack(false, true); err != nil {
					log.Printf("Failed to Nack and requeue message: %v", err)
				} else {
					log.Println("Message Nacked and requeued (NackRequeue)")
				}
			case NackDiscard:
				if err := delivery.Nack(false, false); err != nil {
					log.Printf("Failed to Nack and discard message: %v", err)
				} else {
					log.Println("Message Nacked and discarded (NackDiscard)")
				}
			default:
				log.Printf("Unknown AckType %v, defaulting to Ack", ackType)
				if err := delivery.Ack(false); err != nil {
					log.Printf("Failed to default Ack: %v", err)
				}
			}
		}
	}()

	fmt.Printf("Subscribed to queue '%s' on exchange '%s' with key '%s'\n", queueName, exchange, key)
	return nil
}
func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, func(body []byte) (T, error) {
		var msg T
		return msg, json.Unmarshal(body, &msg)
	})
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, func(body []byte) (T, error) {
		var msg T
		err := gob.NewDecoder(bytes.NewBuffer(body)).Decode(&msg)
		return msg, err
	})
}
