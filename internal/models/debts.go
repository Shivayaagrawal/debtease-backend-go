package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateDebtRequest struct {
	Name               string          `json:"name" validate:"required"`
	Type               string          `json:"type" validate:"required,oneof=credit_card personal_loan student_loan mortgage medical other"`
	Lender             *string         `json:"lender,omitempty"`
	Principal          decimal.Decimal `json:"principal" validate:"required"`
	OutstandingBalance decimal.Decimal `json:"outstanding_balance" validate:"required"`
	InterestRate       decimal.Decimal `json:"interest_rate" validate:"required,gt=0"`
	MinPayment         decimal.Decimal `json:"min_payment" validate:"required,gt=0"`
	DueDate            time.Time       `json:"due_date" validate:"required"` // YYYY-MM-DD
	PaymentFrequency   *string         `json:"payment_frequency,omitempty" validate:"omitempty,oneof=monthly biweekly weekly"`
	DebtTaken          *time.Time      `json:"debt_taken,omitempty"`
	RiskLevel          *string         `json:"risk_level,omitempty" validate:"omitempty,oneof=low medium high"`
	Priority           *int16          `json:"priority,omitempty"`
	Notes              *string         `json:"notes,omitempty"`

	// Credit Card specific fields (required if type=credit_card)
	BillingDay        *int32           `json:"billing_day,omitempty" validate:"omitempty,min=1,max=31"`
	GracePeriodDays   *int32           `json:"grace_period_days,omitempty" validate:"omitempty,min=0"`
	MinPaymentPercent *decimal.Decimal `json:"min_payment_percent,omitempty" validate:"omitempty,gt=0"`

	// Loan specific fields (required if type=personal_loan/student_loan/mortgage)
	PaymentDueDay   *int32           `json:"payment_due_day,omitempty" validate:"omitempty,min=1,max=31"`
	EMIAmount       *decimal.Decimal `json:"emi_amount,omitempty"`
	TenureMonths    *int32           `json:"tenure_months,omitempty" validate:"omitempty,gt=0"`
	MoratoriumUntil *time.Time       `json:"moratorium_until,omitempty"`
}

type UpdateDebtBalanceRequest struct {
	OutstandingBalance decimal.Decimal `json:"outstanding_balance" validate:"required"`
	IsOverdue          bool            `json:"is_overdue"`
}

type UpdateDebtHealthRequest struct {
	DebtHealthScore *decimal.Decimal `json:"debt_health_score" validate:"required"`
}
type DebtResponse struct {
	ID                 string           `json:"id"`
	UserID             string           `json:"user_id"`
	Name               string           `json:"name"`
	Type               string           `json:"type"`
	Lender             *string          `json:"lender,omitempty"`
	Principal          decimal.Decimal  `json:"principal"`
	OutstandingBalance decimal.Decimal  `json:"outstanding_balance"`
	InterestRate       decimal.Decimal  `json:"interest_rate"`
	MinPayment         decimal.Decimal  `json:"min_payment"`
	DueDate            string           `json:"due_date"` // ISO date only
	PaymentFrequency   *string          `json:"payment_frequency,omitempty"`
	IsOverdue          bool             `json:"is_overdue"`
	LateFees           *decimal.Decimal `json:"late_fees,omitempty"`
	OtherCharges       *string          `json:"other_charges,omitempty"`
	DebtTaken          *time.Time       `json:"debt_taken,omitempty"`
	RiskLevel          *string          `json:"risk_level,omitempty"`
	DebtHealthScore    *decimal.Decimal `json:"debt_health_score,omitempty"`
	Priority           *int16           `json:"priority,omitempty"`
	Notes              *string          `json:"notes,omitempty"`
	AccruedInterest    decimal.Decimal  `json:"accrued_interest"`
	LastAccrualDate    time.Time        `json:"last_accrual_date"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`

	// Loan fields (populated if type is loan)
	PaymentDueDay   *int32           `json:"payment_due_day,omitempty"`
	EMIAmount       *decimal.Decimal `json:"emi_amount,omitempty"`
	TenureMonths    *int32           `json:"tenure_months,omitempty"`
	MonthsPaid      *int32           `json:"months_paid,omitempty"`
	MoratoriumUntil *time.Time       `json:"moratorium_until,omitempty"`
	InMoratorium    *bool            `json:"in_moratorium,omitempty"`

	// Credit card fields (populated if type is credit_card)
	CreditCard *CreditCardResponse `json:"credit_card,omitempty"`
}

type CreditCardResponse struct {
	BillingDay        int32           `json:"billing_day"`
	GracePeriodDays   int32           `json:"grace_period_days"`
	MinPaymentPercent decimal.Decimal `json:"min_payment_percent"`
	LastStatementDate *time.Time      `json:"last_statement_date,omitempty"`
	PaymentDueDate    *time.Time      `json:"payment_due_date,omitempty"`
	BilledBalance     decimal.Decimal `json:"billed_balance"`
	UnbilledBalance   decimal.Decimal `json:"unbilled_balance"`
	InGracePeriod     bool            `json:"in_grace_period"`
}
