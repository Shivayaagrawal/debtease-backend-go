package pubsub

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn      *amqp.Connection
	connMutex sync.RWMutex
)

// GetConnection returns a singleton RabbitMQ connection
func GetConnection(rabbitmqURL string) (*amqp.Connection, error) {
	connMutex.RLock()
	if conn != nil && !conn.IsClosed() {
		connMutex.RUnlock()
		return conn, nil
	}
	connMutex.RUnlock()

	connMutex.Lock()
	defer connMutex.Unlock()

	// Double-check after acquiring write lock
	if conn != nil && !conn.IsClosed() {
		return conn, nil
	}

	newConn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	conn = newConn
	return conn, nil
}

// CloseConnection closes the singleton connection (for graceful shutdown)
func CloseConnection() error {
	connMutex.Lock()
	defer connMutex.Unlock()

	if conn != nil && !conn.IsClosed() {
		err := conn.Close()
		conn = nil
		return err
	}
	return nil
}