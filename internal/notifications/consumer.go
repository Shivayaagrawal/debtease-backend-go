package notifications

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"DebtEase/internal/database"
	"DebtEase/internal/events"
	"DebtEase/internal/logger"
	"DebtEase/internal/pubsub"
)

// StartConsumer starts the notification consumer worker
func StartConsumer(ctx context.Context, conn *amqp.Connection, dbQueries *database.Queries) error {
	// Initialize email service
	emailService, err := NewEmailService()
	if err != nil {
		logger.Logger.Warnw("Email service not available (notifications will fail)", "error", err)
		// Continue anyway - we can still log events even if email fails
		emailService = nil
	}

	// Subscribe to user registered events
	err = pubsub.SubscribeJSON(
		conn,
		events.ExchangeNotifications,
		events.QueueUserRegisteredEmailNotifier,
		events.RoutingUserRegistered,
		pubsub.QueueTypeDurable,
		func(event events.UserRegisteredEvent) pubsub.AckType {
			// Create a context with timeout for each message processing
			msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if emailService == nil {
				logger.Logger.Errorw("Email service not available, cannot process user registered event",
					"user_id", event.UserID,
					"email", event.Email)
				// Still acknowledge to avoid infinite requeue
				return pubsub.Ack
			}

			err := HandleUserRegisteredEvent(msgCtx, event, emailService, dbQueries)
			if err != nil {
				logger.Logger.Errorw("Failed to handle user registered event",
					"error", err,
					"user_id", event.UserID,
					"email", event.Email)
				// Requeue on transient errors (you might want to add retry logic here)
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		},
	)

	if err != nil {
		return fmt.Errorf("failed to start notification consumer: %w", err)
	}

	logger.Logger.Info("Notification consumer started successfully")
	return nil
}