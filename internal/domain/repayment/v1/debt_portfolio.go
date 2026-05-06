package repayment_v1

func GeneratePortfolioSchedule(loans []Loan) map[int][]EMIRecord {
	portfolio := make(map[int][]EMIRecord)

	for _, loan := range loans {
		schedule := GenerateLoanSchedule(loan)
		for _, r := range schedule {
			portfolio[r.Month] = append(portfolio[r.Month], r)
		}
	}

	return portfolio
}
