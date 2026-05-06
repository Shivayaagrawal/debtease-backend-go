// internal/pubsub/declare.go
package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// DeclareAndBind creates a channel, declares a queue, binds it to an exchange,
// and returns the channel + queue. The queue is durable or transient based on queueType.
func DeclareAndBind(
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType QueueType,
) (*amqp.Channel, amqp.Queue, error) {

	// 1. Open a new channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// 2. Determine queue flags
	durable := queueType == QueueTypeDurable
	autoDelete := queueType == QueueTypeTransient
	exclusive := queueType == QueueTypeTransient

	// 3. Declare the queue
	args:= amqp.Table{
		"x-dead-letter-exchange": "debtease_dlx",
	}
	
	q, err := ch.QueueDeclare(
		queueName,   // name
		durable,     // durable
		autoDelete,  // auto-delete
		exclusive,   // exclusive
		false,       // no-wait
		args,         // arguments
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	// 4. Bind queue to exchange
	err = ch.QueueBind(
		q.Name,    // queue
		key,       // routing key
		exchange,  // exchange
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}

