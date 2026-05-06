package repayment_v1

import (
	"math"

	"github.com/shopspring/decimal"
)

var (
	decimalZero = decimal.Zero
	twoDecScale = int32(2)
)

// roundHalfUp rounds a decimal to the given scale using round-half-up (away from zero).
func roundHalfUp(d decimal.Decimal, scale int32) decimal.Decimal {
	return d.Round(int32(scale))
}

// money rounds to two decimal places using round-half-up.
func money(d decimal.Decimal) decimal.Decimal {
	return roundHalfUp(d, twoDecScale)
}

// monthlyRate computes monthly rate from annual percent rate (e.g. 12 -> 0.01).
// annualRatePct should be like 12 for 12%.
func monthlyRate(annualRatePct decimal.Decimal) decimal.Decimal {
	// r = annualRate / 12 / 100
	return annualRatePct.Div(decimal.NewFromInt(12)).Div(decimal.NewFromInt(100))
}

// isEffectivelyZero checks if value is within half-paise of zero.
func isEffectivelyZero(d decimal.Decimal) bool {
	abs := d.Abs()
	threshold := decimal.NewFromFloat(0.005) // half of 0.01
	return abs.LessThan(threshold)
}

// powDecimal computes x^y where y is an integer exponent using decimal arithmetics.
// For EMI, tenureMonths is always integer, so this is sufficient.
func powDecimal(base decimal.Decimal, exp int) decimal.Decimal {
	if exp == 0 {
		return decimal.NewFromInt(1)
	}
	if exp < 0 {
		// negative exponents: 1 / (base^|exp|)
		return decimal.NewFromInt(1).Div(powDecimal(base, -exp))
	}
	result := decimal.NewFromInt(1)
	for i := 0; i < exp; i++ {
		result = result.Mul(base)
	}
	return result
}

// toDecimal converts a float64 to decimal with best-effort; used only at API boundary.
func toDecimal(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// fromDecimal converts decimal to float64 for external callers that still expect float64.
// This should be used only in tests or compatibility layers, not for internal money math.
func fromDecimal(d decimal.Decimal) float64 {
	v, _ := d.Float64()
	// guard against NaN/Inf, though decimal shouldn't produce these
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
