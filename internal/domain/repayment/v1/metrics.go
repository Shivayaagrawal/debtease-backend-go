package repayment_v1

import "github.com/shopspring/decimal"

func SummarizePortfolio(schedule map[int][]EMIRecord) PortfolioSummary {
	var paid, interest, principal decimal.Decimal
	lastMonth := 0

	for m, records := range schedule {
		if m > lastMonth {
			lastMonth = m
		}
		for _, r := range records {
			paid = paid.Add(r.EMI)
			interest = interest.Add(r.Interest)
			principal = principal.Add(r.Principal)
		}
	}

	return PortfolioSummary{
		Months:         lastMonth,
		TotalPaid:      money(paid),
		TotalInterest:  money(interest),
		TotalPrincipal: money(principal),
	}
}

