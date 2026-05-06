package service

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"

	"github.com/google/uuid"
)

func GetOrCreateSession(ctx context.Context, db *database.Queries, sessionID uuid.UUID, engineVersion string) (uuid.UUID, error) {
	if sessionID == uuid.Nil {
		// Generate new session ID
		sessionID = uuid.New()
	}

	// Check if session exists
	exists, err := db.SessionExists(ctx, sessionID)
	if err != nil {
		logger.Logger.Errorw("Failed to check session existence", "error", err)
		return uuid.Nil, err
	}

	if !exists {
		// Create new session
		if engineVersion == "" {
			engineVersion = "v1"
		}
		_, err := db.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
			SessionID:     sessionID,
			EngineVersion: engineVersion,
		})
		if err != nil {
			logger.Logger.Errorw("Failed to create calculation session", "error", err, "session_id", sessionID)
			return uuid.Nil, err
		}
		logger.Logger.Infow("Created new calculation session", "session_id", sessionID, "engine_version", engineVersion)
	}

	return sessionID, nil
}

func SessionExists(ctx context.Context, db *database.Queries, sessionID uuid.UUID) (bool, error) {
	return db.SessionExists(ctx, sessionID)
}
