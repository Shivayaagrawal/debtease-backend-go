package service

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
)

// AggregateSessions processes calculation events into session metrics
func AggregateSessions(ctx context.Context, db *database.Queries) error {
	if err := db.AggregateSessions(ctx); err != nil {
		logger.Logger.Errorw("Failed to aggregate sessions", "error", err)
		return err
	}
	logger.Logger.Infow("Successfully aggregated sessions")
	return nil
}

// GetRepeatCalculationRate returns the rate of sessions with 2+ calculation runs
func GetRepeatCalculationRate(ctx context.Context, db *database.Queries) (float64, error) {
	rate, err := db.RepeatCalculationRate(ctx)
	if err != nil {
		logger.Logger.Errorw("Failed to get repeat calculation rate", "error", err)
		return 0, err
	}
	// Convert int32 to float64 (sqlc type mismatch - SQL returns float but Go has int32)
	return float64(rate), nil
}

// GetStrategySwitchRate returns the rate of sessions that switched strategies
func GetStrategySwitchRate(ctx context.Context, db *database.Queries) (float64, error) {
	rate, err := db.StrategySwitchRate(ctx)
	if err != nil {
		logger.Logger.Errorw("Failed to get strategy switch rate", "error", err)
		return 0, err
	}
	// Convert int32 to float64 (sqlc type mismatch - SQL returns float but Go has int32)
	return float64(rate), nil
}

// GetPdfConversionRate returns the rate of sessions that requested PDF
func GetPdfConversionRate(ctx context.Context, db *database.Queries) (float64, error) {
	rate, err := db.PdfConversionRate(ctx)
	if err != nil {
		logger.Logger.Errorw("Failed to get PDF conversion rate", "error", err)
		return 0, err
	}
	// Convert int32 to float64 (sqlc type mismatch - SQL returns float but Go has int32)
	return float64(rate), nil
}

// GetCommitmentRate returns the rate of sessions that both requested PDF and submitted email
func GetCommitmentRate(ctx context.Context, db *database.Queries) (float64, error) {
	rate, err := db.CommitmentRate(ctx)
	if err != nil {
		logger.Logger.Errorw("Failed to get commitment rate", "error", err)
		return 0, err
	}
	// Convert int32 to float64 (sqlc type mismatch - SQL returns float but Go has int32)
	return float64(rate), nil
}

// GetReturningSessionsRate returns the rate of sessions that returned after 1 day
func GetReturningSessionsRate(ctx context.Context, db *database.Queries) (float64, error) {
	rate, err := db.ReturningSessions(ctx)
	if err != nil {
		logger.Logger.Errorw("Failed to get returning sessions rate", "error", err)
		return 0, err
	}
	// Convert int32 to float64 (sqlc type mismatch - SQL returns float but Go has int32)
	return float64(rate), nil
}

// CleanupOldEvents removes calculation events older than 30 days
func CleanupOldEvents(ctx context.Context, db *database.Queries) error {
	if err := db.CleanupOldEvents(ctx); err != nil {
		logger.Logger.Errorw("Failed to cleanup old events", "error", err)
		return err
	}
	logger.Logger.Infow("Successfully cleaned up old events")
	return nil
}
