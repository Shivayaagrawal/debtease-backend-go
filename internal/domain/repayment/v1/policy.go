package repayment_v1

const EngineVersion = "v1.1.0"

var Policy = []string{
	"Fixed interest rate for entire tenure (no reset or floating linkage within this engine version).",
	"Interest is computed on a monthly compounding basis using a monthly rate = (annual_rate_percent / 12 / 100).",
	"All monetary values (principal, interest, EMI, balances, prepayments) are represented in decimal.Decimal, not float, to avoid precision errors.",
	"All monetary amounts are rounded to two decimal places using round-half-up (0.5 is rounded away from zero, not banker’s rounding).",
	"Standard EMI is calculated using the usual amortisation formula P * r * (1+r)^n / ((1+r)^n - 1), using decimal arithmetic and round-half-up.",
	"During moratorium months, no EMI is charged, but interest continues to accrue on the opening balance; opening balance of a month equals the previous month’s closing balance.",
	"During moratorium, accrued interest is added to the principal outstanding (capitalisation of interest) unless the calling system treats it separately; no partial interest waivers are assumed.",
	"Post-moratorium, EMIs are charged every month until the loan is fully closed, following the computed standard EMI unless prepayments/strategies alter the effective payment.",
	"The last instalment is adjusted so that: (a) it never exceeds the normal EMI, (b) principal is fully cleared, and (c) any residual balance within a small tolerance is set to zero.",
	"Prepayments (extra payments above EMI) directly reduce outstanding principal in the same month and reduce future interest; no penalties or charges on prepayments are applied in this engine.",
	"StrategyBaseline pays only scheduled EMIs (no extra); StrategySnowball and StrategyAvalanche allocate monthly extra towards the smallest-balance or highest-rate active loan respectively.",
	"Outstanding principal at start is treated as source of truth from upstream systems; this engine does not reconstruct or reconcile historical schedules.",
	"No bank fees, penalties, taxes, insurance premia, or other non-interest charges are modelled in this engine version.",
	"No floating, benchmark-linked, or step-up/step-down interest rate structures are modelled; rate is constant for the simulated tenure.",
	"Comparisons for loan closure use a small epsilon tolerance (effective zero) to avoid ghost EMIs from sub-paise residuals; balances within tolerance are set to exactly zero.",
}
