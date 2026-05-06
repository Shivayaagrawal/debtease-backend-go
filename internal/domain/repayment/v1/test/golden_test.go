package repayment_v1

import (
	"testing"

	repayment_v1 "DebtEase/internal/domain/repayment/v1"
	"github.com/shopspring/decimal"
)

func TestGoldenCase_SingleLoanBaseline(t *testing.T) {
	t.Parallel()

	loan := repayment_v1.Loan{
		ID:           "HL",
		Principal:    decimal.RequireFromString("1000000"),
		AnnualRate:   decimal.RequireFromString("9"),
		TenureMonths: 120,
	}

	schedule := repayment_v1.GeneratePortfolioSchedule([]repayment_v1.Loan{loan})
	summary := repayment_v1.SummarizePortfolio(schedule)

	// 🔒 DO NOT CHANGE WITHOUT NEW ENGINE VERSION
	if summary.Months != 120 {
		t.Fatalf("months changed: %d", summary.Months)
	}

	// With decimal + round-half-up, expected interest may differ from legacy float engine.
	expectedInterest := decimal.RequireFromString("520109.10")
	diff := summary.TotalInterest.Sub(expectedInterest).Abs()
	if diff.GreaterThan(decimal.NewFromFloat(1)) {
		t.Fatalf(
			"interest drift detected: expected %s got %s",
			expectedInterest.StringFixed(2),
			summary.TotalInterest.StringFixed(2),
		)
	}
}

