package events

import (
	"fmt"

	"DebtEase/internal/logger"
	"DebtEase/internal/pubsub"
)

// PublishEvent publishes an event to RabbitMQ
// Returns error if publishing fails, but logs warnings instead of failing the request
func PublishEvent(rabbitmqURL string, exchange, routingKey string, event interface{}) error {
	if rabbitmqURL == "" {
		logger.Logger.Warnw("RabbitMQ URL not configured, skipping event publish", "event_type", routingKey)
		return nil
	}

	conn, err := pubsub.GetConnection(rabbitmqURL)
	if err != nil {
		logger.Logger.Warnw("Failed to get RabbitMQ connection for event publish", 
			"error", err, 
			"event_type", routingKey)
		return fmt.Errorf("failed to get connection: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Logger.Warnw("Failed to open channel for event publish", 
			"error", err, 
			"event_type", routingKey)
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	err = pubsub.PublishJSON(ch, exchange, routingKey, event)
	if err != nil {
		logger.Logger.Errorw("Failed to publish event", 
			"error", err, 
			"event_type", routingKey,
			"exchange", exchange)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	logger.Logger.Infow("Event published successfully", 
		"event_type", routingKey,
		"exchange", exchange)
	return nil
}