package repayment_v1

import (
	"testing"

	repayment_v1 "DebtEase/internal/domain/repayment/v1"
	"github.com/shopspring/decimal"
)

func TestMoneyConservation(t *testing.T) {
	t.Parallel()

	loans := []repayment_v1.Loan{
		{
			ID:           "HL",
			Principal:    decimal.RequireFromString("1000000"),
			AnnualRate:   decimal.RequireFromString("9"),
			TenureMonths: 120,
		},
	}

	schedule := repayment_v1.GeneratePortfolioSchedule(loans)
	summary := repayment_v1.SummarizePortfolio(schedule)

	repayment_v1.AssertMoneyConservation(summary)
}

func TestNoNegativeBalancesInvariant(t *testing.T) {
	t.Parallel()

	loans := []repayment_v1.Loan{
		{
			ID:           "PL",
			Principal:    decimal.RequireFromString("500000"),
			AnnualRate:   decimal.RequireFromString("12"),
			TenureMonths: 60,
		},
	}

	schedule := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategySnowball, decimal.RequireFromString("20000"))

	repayment_v1.AssertNoNegativeBalances(schedule)
}

