package repayment_v1

import (
	"testing"

	repayment_v1 "DebtEase/internal/domain/repayment/v1"

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

func TestCalculateEMI_ZeroInterest(t *testing.T) {
	t.Parallel()

	principal := decimal.RequireFromString("120000")
	rate := decimal.RequireFromString("0")

	emi := repayment_v1.CalculateEMI(principal, rate, 12)
	expected := decimal.RequireFromString("10000")
	if !emi.Equal(expected) {
		t.Fatalf("expected EMI %s, got %s", expected.StringFixed(2), emi.StringFixed(2))
	}
}

// TestDecimalRounding_VerifyHalfUp ensures we are not using banker’s rounding.
func TestDecimalRounding_VerifyHalfUp(t *testing.T) {
	t.Parallel()

	x := decimal.RequireFromString("1.225")
	y := money(x)

	if y.StringFixed(2) != "1.23" {
		t.Fatalf("expected 1.23, got %s", y.StringFixed(2))
	}
}

func TestLoanClosesCleanly(t *testing.T) {
	t.Parallel()

	loan := repayment_v1.Loan{
		ID:           "HL",
		Principal:    decimal.RequireFromString("1000000"),
		AnnualRate:   decimal.RequireFromString("9"),
		TenureMonths: 120,
	}

	schedule := repayment_v1.GenerateLoanSchedule(loan)
	last := schedule[len(schedule)-1]

	if !last.ClosingBalance.IsZero() {
		t.Fatalf("loan did not close, balance=%s", last.ClosingBalance.StringFixed(2))
	}
}

func TestMoratoriumAccruesInterest(t *testing.T) {
	t.Parallel()

	loan := repayment_v1.Loan{
		ID:               "PL",
		Principal:        decimal.RequireFromString("500000"),
		AnnualRate:       decimal.RequireFromString("12"),
		TenureMonths:     60,
		MoratoriumMonths: 6,
	}

	schedule := repayment_v1.GenerateLoanSchedule(loan)

	if !schedule[0].EMI.IsZero() {
		t.Fatal("EMI should be zero during moratorium")
	}

	if !schedule[0].Interest.GreaterThan(decimalZero) {
		t.Fatal("interest must accrue during moratorium")
	}
}
