package service

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"

	"github.com/google/uuid"
)

func CreateLeadEmail(ctx context.Context, db *database.Queries, email string, source string, sessionID uuid.UUID) error {
	var nullSessionID uuid.NullUUID
	if sessionID != uuid.Nil {
		nullSessionID = uuid.NullUUID{
			UUID:  sessionID,
			Valid: true,
		}
	}

	_, err := db.CreateLeadEmail(ctx, database.CreateLeadEmailParams{
		Email:     email,
		Source:    source,
		SessionID: nullSessionID,
	})
	if err != nil {
		logger.Logger.Errorw("Failed to create lead email", "error", err, "email", email, "source", source)
		return err
	}

	logger.Logger.Infow("Created lead email", "email", email, "source", source, "session_id", sessionID)
	return nil
}

func CheckEmailExists(ctx context.Context, db *database.Queries, email string) (bool, error) {
	exists, err := db.CheckEmailExists(ctx, email)
	if err != nil {
		logger.Logger.Errorw("Failed to check email existence", "error", err, "email", email)
		return false, err
	}
	return exists, nil
}
