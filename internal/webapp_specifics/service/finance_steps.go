package service

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func isNoRows(err error) bool {
	return err != nil && errors.Is(err, sql.ErrNoRows)
}

// GetFinanceSteps loads finance steps for a session. Returns nil, nil if not found.
func GetFinanceSteps(ctx context.Context, db *database.Queries, sessionID uuid.UUID) (*webapp_models.FinanceStepsGetResponse, error) {
	row, err := db.GetFinanceStepsBySessionID(ctx, sessionID)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Logger.Errorw("Failed to get finance steps", "error", err, "session_id", sessionID)
		return nil, err
	}
	return financeStepToGetResponse(row), nil
}

// UpsertFinanceSteps merges request into existing (if any) and saves. Session must already exist.
func UpsertFinanceSteps(ctx context.Context, db *database.Queries, sessionID uuid.UUID, req *webapp_models.FinanceStepsRequest, engineVersion string) (*webapp_models.FinanceStepsPostResponse, error) {
	existing, errExisting := db.GetFinanceStepsBySessionID(ctx, sessionID)
	hasExisting := errExisting == nil

	var monthlyIncome, fixedObligations decimal.NullDecimal
	var loansRaw json.RawMessage = []byte("[]")
	ev := engineVersion
	if ev == "" {
		ev = "v1"
	}

	if req.MonthlyIncome != nil {
		monthlyIncome = decimal.NullDecimal{Decimal: *req.MonthlyIncome, Valid: true}
	} else if hasExisting && existing.MonthlyIncome.Valid {
		monthlyIncome = existing.MonthlyIncome
	}
	if req.FixedObligations != nil {
		fixedObligations = decimal.NullDecimal{Decimal: *req.FixedObligations, Valid: true}
	} else if hasExisting && existing.FixedObligations.Valid {
		fixedObligations = existing.FixedObligations
	}
	if len(req.Loans) > 0 {
		b, _ := json.Marshal(req.Loans)
		loansRaw = b
	} else if hasExisting && len(existing.Loans) > 0 && string(existing.Loans) != "[]" {
		loansRaw = existing.Loans
	}
	if req.EngineVersion != nil && *req.EngineVersion != "" {
		ev = *req.EngineVersion
	} else if hasExisting && existing.EngineVersion.Valid {
		ev = existing.EngineVersion.String
	}

	params := database.UpsertFinanceStepsParams{
		SessionID:         sessionID,
		MonthlyIncome:    monthlyIncome,
		FixedObligations: fixedObligations,
		Loans:             loansRaw,
		EngineVersion:    toNullString(ev),
	}
	row, err := db.UpsertFinanceSteps(ctx, params)
	if err != nil {
		logger.Logger.Errorw("Failed to upsert finance steps", "error", err, "session_id", sessionID)
		return nil, err
	}
	return financeStepToPostResponse(row), nil
}

// ReplaceFinanceSteps full-replace: only request values are saved (no merge). Use for PUT.
func ReplaceFinanceSteps(ctx context.Context, db *database.Queries, sessionID uuid.UUID, req *webapp_models.FinanceStepsRequest, engineVersion string) (*webapp_models.FinanceStepsPostResponse, error) {
	var monthlyIncome, fixedObligations decimal.NullDecimal
	loansRaw := json.RawMessage([]byte("[]"))
	ev := engineVersion
	if ev == "" {
		ev = "v1"
	}
	if req.MonthlyIncome != nil {
		monthlyIncome = decimal.NullDecimal{Decimal: *req.MonthlyIncome, Valid: true}
	}
	if req.FixedObligations != nil {
		fixedObligations = decimal.NullDecimal{Decimal: *req.FixedObligations, Valid: true}
	}
	if len(req.Loans) > 0 {
		b, _ := json.Marshal(req.Loans)
		loansRaw = b
	}
	if req.EngineVersion != nil && *req.EngineVersion != "" {
		ev = *req.EngineVersion
	}
	params := database.UpsertFinanceStepsParams{
		SessionID:         sessionID,
		MonthlyIncome:    monthlyIncome,
		FixedObligations: fixedObligations,
		Loans:             loansRaw,
		EngineVersion:    toNullString(ev),
	}
	row, err := db.UpsertFinanceSteps(ctx, params)
	if err != nil {
		logger.Logger.Errorw("Failed to replace finance steps", "error", err, "session_id", sessionID)
		return nil, err
	}
	return financeStepToPostResponse(row), nil
}

// DeleteFinanceSteps removes saved finance steps for the session.
func DeleteFinanceSteps(ctx context.Context, db *database.Queries, sessionID uuid.UUID) error {
	return db.DeleteFinanceStepsBySessionID(ctx, sessionID)
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func financeStepToGetResponse(row database.FinanceStep) *webapp_models.FinanceStepsGetResponse {
	resp := &webapp_models.FinanceStepsGetResponse{
		SessionID: row.SessionID.String(),
	}
	if row.MonthlyIncome.Valid {
		resp.FinanceSteps.MonthlyIncome = &row.MonthlyIncome.Decimal
	}
	if row.FixedObligations.Valid {
		resp.FinanceSteps.FixedObligations = &row.FixedObligations.Decimal
	}
	if row.EngineVersion.Valid {
		resp.FinanceSteps.EngineVersion = row.EngineVersion.String
	}
	if len(row.Loans) > 0 && string(row.Loans) != "[]" {
		_ = json.Unmarshal(row.Loans, &resp.FinanceSteps.Loans)
	}
	return resp
}

func financeStepToPostResponse(row database.FinanceStep) *webapp_models.FinanceStepsPostResponse {
	resp := &webapp_models.FinanceStepsPostResponse{
		SessionID: row.SessionID.String(),
		SavedAt:   row.UpdatedAt,
	}
	if row.MonthlyIncome.Valid {
		resp.FinanceSteps.MonthlyIncome = &row.MonthlyIncome.Decimal
	}
	if row.FixedObligations.Valid {
		resp.FinanceSteps.FixedObligations = &row.FixedObligations.Decimal
	}
	if row.EngineVersion.Valid {
		resp.FinanceSteps.EngineVersion = row.EngineVersion.String
	}
	if len(row.Loans) > 0 && string(row.Loans) != "[]" {
		_ = json.Unmarshal(row.Loans, &resp.FinanceSteps.Loans)
	}
	return resp
}
