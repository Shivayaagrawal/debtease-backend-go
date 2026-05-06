package repayment_v1

import "github.com/shopspring/decimal"

type loanState struct {
	Loan    Loan
	Balance decimal.Decimal
	EMI     decimal.Decimal
	Active  bool
}

func ApplyStrategy(
	loans []Loan,
	strategy StrategyType,
	monthlyExtra decimal.Decimal,
) map[int][]EMIRecord {

	states := make(map[string]*loanState)
	for _, l := range loans {
		states[l.ID] = &loanState{
			Loan:    l,
			Balance: l.Principal,
			EMI:     CalculateEMI(l.Principal, l.AnnualRate, l.TenureMonths),
			Active:  true,
		}
	}

	portfolio := make(map[int][]EMIRecord)
	month := 1

	for anyActive(states) {
		var records []EMIRecord

		// Normal EMI pass
		for _, s := range states {
			if !s.Active {
				continue
			}

			r := monthlyRate(s.Loan.AnnualRate)
			opening := s.Balance
			interest := money(opening.Mul(r))

			totalDue := money(opening.Add(interest))
			emi := decimal.Min(s.EMI, totalDue)
			principal := money(emi.Sub(interest))
			closing := money(opening.Sub(principal))

			if closing.LessThan(decimalZero) || isEffectivelyZero(closing) {
				closing = decimalZero
				s.Active = false
			}

			s.Balance = closing

			records = append(records, EMIRecord{
				LoanID:         s.Loan.ID,
				Month:          month,
				OpeningBalance: opening,
				EMI:            emi,
				Interest:       interest,
				Principal:      principal,
				ClosingBalance: closing,
			})
		}

		// Prepayment
		if monthlyExtra.GreaterThan(decimalZero) && strategy != StrategyBaseline {
			target := selectTarget(states, strategy)
			if target != nil {
				extra := decimal.Min(monthlyExtra, target.Balance)
				target.Balance = money(target.Balance.Sub(extra))

				for i := range records {
					if records[i].LoanID == target.Loan.ID {
						records[i].Principal = records[i].Principal.Add(extra)
						records[i].ClosingBalance = target.Balance
						break
					}
				}

				if target.Balance.LessThan(decimalZero) || isEffectivelyZero(target.Balance) {
					target.Balance = decimalZero
					target.Active = false
				}
			}
		}

		portfolio[month] = records
		month++
	}

	return portfolio
}

func selectTarget(states map[string]*loanState, strategy StrategyType) *loanState {
	var selected *loanState

	for _, s := range states {
		if !s.Active {
			continue
		}

		if selected == nil {
			selected = s
			continue
		}

		switch strategy {
		case StrategySnowball:
			if s.Balance.LessThan(selected.Balance) {
				selected = s
			}
		case StrategyAvalanche:
			if s.Loan.AnnualRate.GreaterThan(selected.Loan.AnnualRate) {
				selected = s
			}
		}
	}

	return selected
}

func anyActive(states map[string]*loanState) bool {
	for _, s := range states {
		if s.Active {
			return true
		}
	}
	return false
}

