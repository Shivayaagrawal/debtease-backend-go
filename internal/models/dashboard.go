package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type DashboardResponse struct {
	// Total Outstanding Debt
	TotalOutstanding decimal.Decimal `json:"total_outstanding"`

	// Repayment Progress (for ring chart)
	RepaymentProgress RepaymentProgress `json:"repayment_progress"`

	// Debt Health
	DebtHealth DebtHealth `json:"debt_health"`

	// Expected Debt Free
	ExpectedDebtFree *time.Time `json:"expected_debt_free,omitempty"`
	MonthsToDebtFree *int32     `json:"months_to_debt_free,omitempty"`

	// Monthly Payment
	MonthlyMinimumPayment decimal.Decimal `json:"monthly_minimum_payment"`

	// Average Interest
	AverageInterestRate decimal.Decimal `json:"average_interest_rate"`

	// Upcoming EMIs
	UpcomingEMIs []UpcomingEMIResponse `json:"upcoming_emis"`

	// Payment History Summary
	PaymentHistory PaymentHistorySummary `json:"payment_history"`
}

type RepaymentProgress struct {
	TotalPrincipal      decimal.Decimal `json:"total_principal"`
	TotalPaid           decimal.Decimal `json:"total_paid"`
	RepaymentPercentage decimal.Decimal `json:"repayment_percentage"`
}

type DebtHealth struct {
	OverallHealthScore decimal.Decimal `json:"overall_health_score"`
	TotalActiveDebts   int64           `json:"total_active_debts"`
	OverdueCount       int64           `json:"overdue_count"`
	HighRiskCount      int64           `json:"high_risk_count"`
}

type UpcomingEMIResponse struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Type               string           `json:"type"`
	Lender             *string          `json:"lender,omitempty"`
	OutstandingBalance decimal.Decimal  `json:"outstanding_balance"`
	AccruedInterest    decimal.Decimal  `json:"accrued_interest"`
	PaymentAmount      decimal.Decimal  `json:"payment_amount"`
	DueDate            string           `json:"due_date"` // ISO date format
	PaymentDueDay      *int32           `json:"payment_due_day,omitempty"`
	IsOverdue          bool             `json:"is_overdue"`
	LateFees           *decimal.Decimal `json:"late_fees,omitempty"`
	DaysUntilDue       int32            `json:"days_until_due"`
}

type PaymentHistorySummary struct {
	PaidLast30Days     decimal.Decimal `json:"paid_last_30_days"`
	PaidLast90Days     decimal.Decimal `json:"paid_last_90_days"`
	PaidLastYear       decimal.Decimal `json:"paid_last_year"`
	PaymentsLast30Days int64           `json:"payments_last_30_days"`
}
