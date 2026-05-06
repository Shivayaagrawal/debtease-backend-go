package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type RecordPaymentRequest struct {
	Amount        *decimal.Decimal `json:"amount,omitempty" validate:"omitempty,gt=0"`
	PaymentDate   time.Time        `json:"payment_date" validate:"required"`
	Type          string           `json:"type" validate:"required,oneof=minimum full extra"`
	Notes         *string          `json:"notes,omitempty" validate:"omitempty"`
	PaymentMethod *string          `json:"payment_method,omitempty" validate:"omitempty"`
}

type PaymentResponse struct {
	ID               string          `json:"id"`
	DebtID           string          `json:"debt_id"`
	Amount           decimal.Decimal `json:"amount"`
	PaymentDate      string          `json:"payment_date"`
	Type             string          `json:"type"`
	Notes            *string         `json:"notes,omitempty"`
	PrincipalApplied decimal.Decimal `json:"principal_applied"`
	InterestApplied  decimal.Decimal `json:"interest_applied"`
	PaymentMethod    *string         `json:"payment_method,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

