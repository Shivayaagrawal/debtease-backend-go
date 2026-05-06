package repayment_v1

import "github.com/shopspring/decimal"

// CalculateEMI computes the standard EMI for a loan using decimal arithmetic.
// principal and annualRatePct are decimals; annualRatePct is like 12 for 12%.
func CalculateEMI(principal decimal.Decimal, annualRatePct decimal.Decimal, tenureMonths int) decimal.Decimal {
	if tenureMonths <= 0 {
		return decimalZero
	}

	r := monthlyRate(annualRatePct) // decimal monthly rate
	if r.IsZero() {
		return money(principal.Div(decimal.NewFromInt(int64(tenureMonths))))
	}

	one := decimal.NewFromInt(1)
	rPlusOne := one.Add(r)
	pow := powDecimal(rPlusOne, tenureMonths)

	// emi = P * r * (1+r)^n / ((1+r)^n - 1)
	numerator := principal.Mul(r).Mul(pow)
	denominator := pow.Sub(one)
	if denominator.IsZero() {
		return money(principal.Div(decimal.NewFromInt(int64(tenureMonths))))
	}

	emi := numerator.Div(denominator)
	return money(emi)
}

// GenerateLoanSchedule generates the EMI schedule for a single loan.
func GenerateLoanSchedule(loan Loan) []EMIRecord {
	r := monthlyRate(loan.AnnualRate)
	emi := CalculateEMI(loan.Principal, loan.AnnualRate, loan.TenureMonths)
	balance := loan.Principal

	var schedule []EMIRecord
	month := 1

	// Moratorium: interest accrues, no EMI; opening balance is previous closing.
	for i := 0; i < loan.MoratoriumMonths; i++ {
		opening := balance
		interest := money(opening.Mul(r))
		balance = money(opening.Add(interest))

		schedule = append(schedule, EMIRecord{
			LoanID:         loan.ID,
			Month:          month,
			OpeningBalance: opening,
			EMI:            decimalZero,
			Interest:       interest,
			Principal:      decimalZero,
			ClosingBalance: balance,
		})
		month++
	}

	// Repayment
	for !isEffectivelyZero(balance) && balance.GreaterThan(decimalZero) {
		opening := balance
		interest := money(opening.Mul(r))

		// Determine if this is the last installment: total due <= normal EMI within tolerance.
		totalDue := money(opening.Add(interest))
		effectiveEMI := emi
		principal := decimalZero
		closing := decimalZero

		if totalDue.LessThanOrEqual(emi) || totalDue.Sub(emi).Abs().LessThan(decimal.NewFromFloat(0.005)) {
			// Last installment: keep EMI <= normal EMI, clear principal fully.
			effectiveEMI = totalDue
			principal = opening
			closing = decimalZero
		} else {
			// Regular installment.
			effectiveEMI = decimal.Min(emi, totalDue)
			principal = money(effectiveEMI.Sub(interest))
			closing = money(opening.Sub(principal))
			if closing.LessThan(decimalZero) || isEffectivelyZero(closing) {
				closing = decimalZero
			}
		}

		balance = closing

		schedule = append(schedule, EMIRecord{
			LoanID:         loan.ID,
			Month:          month,
			OpeningBalance: opening,
			EMI:            effectiveEMI,
			Interest:       interest,
			Principal:      principal,
			ClosingBalance: closing,
		})

		month++
	}

	return schedule
}
