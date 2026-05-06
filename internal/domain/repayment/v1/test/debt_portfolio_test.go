package repayment_v1

import (
	"testing"

	repayment_v1 "DebtEase/internal/domain/repayment/v1"
	"github.com/shopspring/decimal"
)

func TestPortfolioScheduleAggregatesLoans(t *testing.T) {
	t.Parallel()

	loans := []repayment_v1.Loan{
		{
			ID:           "HL",
			Principal:    decimal.RequireFromString("1000000"),
			AnnualRate:   decimal.RequireFromString("9"),
			TenureMonths: 120,
		},
		{
			ID:           "PL",
			Principal:    decimal.RequireFromString("300000"),
			AnnualRate:   decimal.RequireFromString("12"),
			TenureMonths: 36,
		},
	}

	portfolio := repayment_v1.GeneratePortfolioSchedule(loans)

	if len(portfolio) == 0 {
		t.Fatal("portfolio schedule empty")
	}

	month1 := portfolio[1]
	if len(month1) != 2 {
		t.Fatalf("expected 2 loan records in month 1, got %d", len(month1))
	}
}

