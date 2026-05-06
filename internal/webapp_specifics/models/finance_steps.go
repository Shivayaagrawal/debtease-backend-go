package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// FinanceStepsRequest is the body for POST /api/v1/finance-steps.
type FinanceStepsRequest struct {
	SessionID         *string          `json:"session_id,omitempty"`
	MonthlyIncome     *decimal.Decimal `json:"monthly_income,omitempty"`
	FixedObligations  *decimal.Decimal `json:"fixed_obligations,omitempty"`
	Loans             []LoanInput     `json:"loans,omitempty"`
	EngineVersion     *string          `json:"engine_version,omitempty"`
}

// FinanceStepsPayload is the stored finance steps (for response).
type FinanceStepsPayload struct {
	MonthlyIncome    *decimal.Decimal `json:"monthly_income,omitempty"`
	FixedObligations *decimal.Decimal `json:"fixed_obligations,omitempty"`
	Loans            []LoanInput     `json:"loans,omitempty"`
	EngineVersion    string          `json:"engine_version,omitempty"`
}

// FinanceStepsPostResponse is the response for POST /api/v1/finance-steps.
type FinanceStepsPostResponse struct {
	SessionID   string             `json:"session_id"`
	SavedAt     time.Time          `json:"saved_at"`
	FinanceSteps FinanceStepsPayload `json:"finance_steps"`
}

// FinanceStepsGetResponse is the response for GET /api/v1/finance-steps.
type FinanceStepsGetResponse struct {
	SessionID    string             `json:"session_id"`
	FinanceSteps FinanceStepsPayload `json:"finance_steps"`
}
