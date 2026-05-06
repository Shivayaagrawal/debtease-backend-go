package repayment_v1

import "github.com/shopspring/decimal"

var eps = decimal.NewFromFloat(0.01)

func AssertMoneyConservation(summary PortfolioSummary) {
	if summary.TotalPaid.Sub(summary.TotalInterest.Add(summary.TotalPrincipal)).Abs().GreaterThan(eps) {
		panic("money conservation violated")
	}
}

func AssertNoNegativeBalances(schedule map[int][]EMIRecord) {
	for _, records := range schedule {
		for _, r := range records {
			if r.ClosingBalance.LessThan(decimalZero) && r.ClosingBalance.Abs().GreaterThan(eps) {
				panic("negative balance detected")
			}
		}
	}
}

