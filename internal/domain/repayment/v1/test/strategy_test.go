package repayment_v1

import (
	"testing"

	repayment_v1 "DebtEase/internal/domain/repayment/v1"
	"github.com/shopspring/decimal"
)

func TestAvalancheBeatsBaseline(t *testing.T) {
	t.Parallel()

	loans := []repayment_v1.Loan{
		{
			ID:           "HL",
			Principal:    decimal.RequireFromString("2000000"),
			AnnualRate:   decimal.RequireFromString("8.5"),
			TenureMonths: 240,
		},
		{
			ID:           "PL",
			Principal:    decimal.RequireFromString("500000"),
			AnnualRate:   decimal.RequireFromString("13"),
			TenureMonths: 60,
		},
	}

	baseline := repayment_v1.GeneratePortfolioSchedule(loans)
	avalanche := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategyAvalanche, decimal.RequireFromString("10000"))

	baseSummary := repayment_v1.SummarizePortfolio(baseline)
	avaSummary := repayment_v1.SummarizePortfolio(avalanche)

	if !avaSummary.TotalInterest.LessThan(baseSummary.TotalInterest) {
		t.Fatal("avalanche should reduce total interest")
	}
}

func TestStrategyDoesNotRewritePast(t *testing.T) {
	t.Parallel()

	loans := []repayment_v1.Loan{
		{
			ID:           "PL",
			Principal:    decimal.RequireFromString("500000"),
			AnnualRate:   decimal.RequireFromString("12"),
			TenureMonths: 60,
		},
	}

	baseline := repayment_v1.GeneratePortfolioSchedule(loans)
	switched := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategySnowball, decimal.RequireFromString("5000"))

	for month := 1; month <= 6; month++ {
		baseEMI := sumEMI(baseline[month])
		switchEMI := sumEMI(switched[month])

		diff := baseEMI.Sub(switchEMI).Abs()
		if diff.GreaterThan(decimal.NewFromFloat(0.01)) {
			t.Fatalf("EMI mismatch in month %d", month)
		}
	}
}

func TestAggressivePrepaymentNeverGoesNegative(t *testing.T) {
	t.Parallel()

	loans := []repayment_v1.Loan{
		{
			ID:           "CL",
			Principal:    decimal.RequireFromString("300000"),
			AnnualRate:   decimal.RequireFromString("10"),
			TenureMonths: 36,
		},
	}

	schedule := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategyAvalanche, decimal.RequireFromString("50000"))

	repayment_v1.AssertNoNegativeBalances(schedule)
}

func sumEMI(records []repayment_v1.EMIRecord) decimal.Decimal {
	total := decimal.Zero
	for _, r := range records {
		total = total.Add(r.EMI)
	}
	return total
}

