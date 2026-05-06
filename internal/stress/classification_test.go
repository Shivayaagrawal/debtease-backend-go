package stress

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestComputeMetrics_EmptyLoans(t *testing.T) {
	t.Parallel()
	m := ComputeMetrics(nil, decimal.Zero, decimal.Zero)
	if m.FragmentationIndex != 0 || m.EMIRatio != 0 {
		t.Errorf("empty loans: got %+v", m)
	}
}

func TestComputeMetrics_OneLoan(t *testing.T) {
	t.Parallel()
	income, _ := decimal.NewFromString("100000")
	loans := []LoanInput{
		{ID: "l1", Principal: decimal.RequireFromString("50000"), AnnualRate: decimal.RequireFromString("12"), TenureMonths: 12, MoratoriumMonths: 0},
	}
	m := ComputeMetrics(loans, income, decimal.Zero)
	if m.FragmentationIndex != 1 {
		t.Errorf("fragmentation_index: got %f, want 1", m.FragmentationIndex)
	}
	if m.EMIRatio <= 0 {
		t.Errorf("emi_ratio should be positive, got %f", m.EMIRatio)
	}
}

func TestClassify_SafeToHold(t *testing.T) {
	t.Parallel()
	m := StressMetrics{EMIRatio: 0.25, FragmentationIndex: 2, ShockBufferMonths: 2, UnsecuredRatio: 0.5}
	c, conf, flags := Classify(m)
	if c != ClassificationSafeToHold {
		t.Errorf("got classification %s, want SAFE_TO_HOLD", c)
	}
	if conf != "high" {
		t.Errorf("got confidence %s", conf)
	}
	if len(flags) != 0 {
		t.Errorf("expected no flags, got %v", flags)
	}
}

func TestClassify_NeedsMonitoring(t *testing.T) {
	t.Parallel()
	m := StressMetrics{EMIRatio: 0.38, FragmentationIndex: 3, ShockBufferMonths: 1.2, UnsecuredRatio: 0.5}
	c, _, _ := Classify(m)
	if c != ClassificationNeedsMonitoring {
		t.Errorf("got classification %s, want NEEDS_MONITORING", c)
	}
}

func TestClassify_AvoidAdding(t *testing.T) {
	t.Parallel()
	m := StressMetrics{EMIRatio: 0.5, FragmentationIndex: 5, ShockBufferMonths: 0.5, UnsecuredRatio: 0.7}
	c, _, flags := Classify(m)
	if c != ClassificationAvoidAdding {
		t.Errorf("got classification %s, want AVOID_ADDING", c)
	}
	if len(flags) == 0 {
		t.Errorf("expected flags for avoid_adding")
	}
}
