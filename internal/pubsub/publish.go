package pubsub

import (
	"context"
	"encoding/json"
	"encoding/gob"
	"bytes"
	amqp "github.com/rabbitmq/amqp091-go"
)



func PublishJSON[T any]( channel *amqp.Channel, exchange, routingKey string, message T) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return channel.PublishWithContext(context.Background(), exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:       body,
	})
}

func PublishGob[T any](channel *amqp.Channel, exchange, routingKey string, message T) error {
    var buf bytes.Buffer
    if err := gob.NewEncoder(&buf).Encode(message); err != nil {
        return err
    }
    return channel.PublishWithContext(
        context.Background(),
        exchange,
        routingKey,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/gob",
            Body:        buf.Bytes(),
        },
    )
}