package service

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidEventType = errors.New("invalid event type")
)

var allowedEventTypes = map[string]bool{
	"calculator_started":          true,
	"result_viewed":               true,
	"strategy_changed":            true,
	"input_modified_after_result": true,
	"pdf_requested":               true,
	"email_submitted":             true,
}

func EmitEvent(ctx context.Context, db *database.Queries, sessionID uuid.UUID, eventType string, metadata map[string]interface{}) error {
	// Validate event type
	if !allowedEventTypes[eventType] {
		logger.Logger.Warnw("Invalid event type attempted", "event_type", eventType, "session_id", sessionID)
		return ErrInvalidEventType
	}

	// Convert metadata to JSONB
	var metadataJSONB json.RawMessage
	if metadata == nil {
		metadataJSONB = []byte("{}")
	} else {
		var err error
		metadataJSONB, err = json.Marshal(metadata)
		if err != nil {
			logger.Logger.Errorw("Failed to marshal event metadata", "error", err)
			return err
		}
	}

	_, err := db.CreateCalculationEvent(ctx, database.CreateCalculationEventParams{
		SessionID: sessionID,
		EventType: eventType,
		Metadata:  metadataJSONB,
	})
	if err != nil {
		logger.Logger.Errorw("Failed to create calculation event", "error", err, "session_id", sessionID, "event_type", eventType)
		return err
	}

	logger.Logger.Infow("Emitted calculation event", "session_id", sessionID, "event_type", eventType)
	return nil
}
