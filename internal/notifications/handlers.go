package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sqlc-dev/pqtype"
	"DebtEase/internal/database"
	"DebtEase/internal/events"
	"DebtEase/internal/logger"
)

// HandleUserRegisteredEvent processes user registration events and sends welcome email
func HandleUserRegisteredEvent(
	ctx context.Context,
	event events.UserRegisteredEvent,
	emailService *EmailService,
	dbQueries *database.Queries,
) error {
	// Create notification event record with pending status
	metadataJSON, err := json.Marshal(map[string]interface{}{
		"name":       event.Name,
		"email":      event.Email,
		"created_at": event.CreatedAt,
	})
	if err != nil {
		logger.Logger.Errorw("Failed to marshal metadata", "error", err)
		metadataJSON = []byte("{}")
	}

	notificationEvent, err := dbQueries.CreateNotificationEvent(ctx, database.CreateNotificationEventParams{
		UserID:    event.UserID,
		EventType: events.EventTypeUserRegistered,
		Channel:   ChannelEmail,
		Recipient: event.Email,
		Status:    "pending",
		Metadata:  pqtype.NullRawMessage{RawMessage: metadataJSON, Valid: true},
	})
	if err != nil {
		logger.Logger.Errorw("Failed to create notification event record", "error", err, "user_id", event.UserID)
		return fmt.Errorf("failed to create notification event: %w", err)
	}

	// Send welcome email
	err = emailService.SendWelcomeEmail(ctx, event.Email, event.Name)
	if err != nil {
		// Update status to failed
		errorMsg := sql.NullString{String: err.Error(), Valid: true}
		_, updateErr := dbQueries.UpdateNotificationEventStatus(ctx, database.UpdateNotificationEventStatusParams{
			ID:           notificationEvent.ID,
			Status:       "failed",
			ErrorMessage: errorMsg,
		})
		if updateErr != nil {
			logger.Logger.Errorw("Failed to update notification event status", "error", updateErr)
		}
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	// Update status to sent
	_, err = dbQueries.UpdateNotificationEventStatus(ctx, database.UpdateNotificationEventStatusParams{
		ID:           notificationEvent.ID,
		Status:       "sent",
		ErrorMessage: sql.NullString{Valid: false},
	})
	if err != nil {
		logger.Logger.Errorw("Failed to update notification event status to sent", "error", err)
		// Don't return error - email was sent successfully
	}

	logger.Logger.Infow("User registered event processed successfully", 
		"user_id", event.UserID, 
		"email", event.Email,
		"notification_id", notificationEvent.ID)
	
	return nil
}