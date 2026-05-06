package stress

import (
	repayment_v1 "DebtEase/internal/domain/repayment/v1"
	"github.com/shopspring/decimal"
)

// Policy thresholds (India v1) from loan_classifications_user_risk.md.
const (
	MaxSafeEMIRatio        = 0.35
	MaxFragmentationIndex  = 4.0
	MinShockBufferMonths   = 1.5
	MaxUnsecuredRatio     = 0.6
)

// ComputeMetrics computes EMI ratio, fragmentation, rigidity, shock buffer, unsecured ratio.
// monthlyIncome and fixedObligations can be zero if not provided.
func ComputeMetrics(loans []LoanInput, monthlyIncome, fixedObligations decimal.Decimal) StressMetrics {
	m := StressMetrics{}
	if len(loans) == 0 {
		return m
	}

	totalEMI := decimal.Zero
	unsecuredEMI := decimal.Zero
	rigiditySum := decimal.Zero

	for _, loan := range loans {
		emi := repayment_v1.CalculateEMI(loan.Principal, loan.AnnualRate, loan.TenureMonths)
		totalEMI = totalEMI.Add(emi)
		secured := loan.Secured != nil && *loan.Secured
		if !secured {
			unsecuredEMI = unsecuredEMI.Add(emi)
		}
		rigiditySum = rigiditySum.Add(emi.Mul(decimal.NewFromInt(int64(loan.TenureMonths))))
	}

	m.FragmentationIndex = float64(len(loans)) // v1: no BNPL weighting

	if !monthlyIncome.IsZero() {
		m.EMIRatio, _ = totalEMI.Div(monthlyIncome).Float64()
		if !rigiditySum.IsZero() {
			m.RigidityScore, _ = rigiditySum.Div(monthlyIncome).Float64()
		}
		if !totalEMI.IsZero() {
			netAfterFixed := monthlyIncome.Sub(fixedObligations).Sub(totalEMI)
			m.ShockBufferMonths, _ = netAfterFixed.Div(totalEMI).Float64()
		}
	}
	if !totalEMI.IsZero() && !unsecuredEMI.IsZero() {
		m.UnsecuredRatio, _ = unsecuredEMI.Div(totalEMI).Float64()
	} else if !totalEMI.IsZero() {
		m.UnsecuredRatio = 1.0
	}
	return m
}
