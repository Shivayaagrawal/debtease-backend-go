package repayment_v1

import "github.com/shopspring/decimal"

type StrategyType string

const (
	StrategyBaseline  StrategyType = "BASELINE"
	StrategySnowball  StrategyType = "SNOWBALL"
	StrategyAvalanche StrategyType = "AVALANCHE"
)

// Loan represents a single loan's basic parameters.
// Money values use decimal.Decimal to avoid floating point rounding errors.
type Loan struct {
	ID               string
	Principal        decimal.Decimal
	AnnualRate       decimal.Decimal // percentage, e.g. 12 for 12%
	TenureMonths     int
	MoratoriumMonths int
}

// EMIRecord represents a single period's cashflow for a loan.
type EMIRecord struct {
	LoanID         string
	Month          int
	OpeningBalance decimal.Decimal
	EMI            decimal.Decimal
	Interest       decimal.Decimal
	Principal      decimal.Decimal
	ClosingBalance decimal.Decimal
}

type PortfolioSummary struct {
	Months         int
	TotalPaid      decimal.Decimal
	TotalInterest  decimal.Decimal
	TotalPrincipal decimal.Decimal
}

type CalculationInput struct {
	Loans        []Loan
	Strategy     StrategyType
	MonthlyExtra decimal.Decimal
}

type CalculationResult struct {
	Schedule      map[int][]EMIRecord
	Summary       PortfolioSummary
	EngineVersion string
}
