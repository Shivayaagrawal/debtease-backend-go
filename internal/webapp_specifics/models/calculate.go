package models

import "github.com/shopspring/decimal"

type LoanInput struct {
	ID               string          `json:"id" validate:"required"`
	Principal        decimal.Decimal `json:"principal" validate:"required,gt=0"`
	AnnualRate       decimal.Decimal `json:"annual_rate" validate:"required,gte=0"`
	TenureMonths     int             `json:"tenure_months" validate:"required,gt=0"`
	MoratoriumMonths int             `json:"moratorium_months" validate:"gte=0"`
}

type BaselineCalculateRequest struct {
	EngineVersion string      `json:"engine_version"`
	Loans         []LoanInput `json:"loans" validate:"required,min=1"`
}

type StrategyCalculateRequest struct {
	EngineVersion string          `json:"engine_version" validate:"required"`
	Strategy      string          `json:"strategy" validate:"required,oneof=AVALANCHE SNOWBALL"`
	MonthlyExtra  decimal.Decimal `json:"monthly_extra" validate:"gte=0"`
	Loans         []LoanInput     `json:"loans" validate:"required,min=1"`
}

type PortfolioSummary struct {
	Months         int             `json:"months"`
	TotalPaid      decimal.Decimal `json:"total_paid"`
	TotalInterest  decimal.Decimal `json:"total_interest"`
	TotalPrincipal decimal.Decimal `json:"total_principal"`
}

type EMIRecord struct {
	LoanID         string          `json:"loan_id"`
	Month          int             `json:"month"`
	OpeningBalance decimal.Decimal `json:"opening_balance"`
	EMI            decimal.Decimal `json:"emi"`
	Interest       decimal.Decimal `json:"interest"`
	Principal      decimal.Decimal `json:"principal"`
	ClosingBalance decimal.Decimal `json:"closing_balance"`
}

type BaselineCalculateResponse struct {
	EngineVersion string              `json:"engine_version"`
	Summary       PortfolioSummary    `json:"summary"`
	Schedule      map[int][]EMIRecord `json:"schedule"`
}

type Delta struct {
	InterestSaved decimal.Decimal `json:"interest_saved"`
	MonthsSaved   int             `json:"months_saved"`
}

type StrategyCalculateResponse struct {
	EngineVersion   string              `json:"engine_version"`
	Strategy        string              `json:"strategy"`
	BaselineSummary PortfolioSummary    `json:"baseline_summary"`
	StrategySummary PortfolioSummary    `json:"strategy_summary"`
	Delta           Delta               `json:"delta"`
	Schedule        map[int][]EMIRecord `json:"schedule"`
}
