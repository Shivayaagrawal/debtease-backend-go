package stress

import (
	"github.com/shopspring/decimal"
)

// ClassificationRequest is the body for POST /api/v1/classification.
type ClassificationRequest struct {
	Loans             []LoanInput     `json:"loans"`
	MonthlyIncome     *decimal.Decimal `json:"monthly_income,omitempty"`
	FixedObligations  *decimal.Decimal `json:"fixed_obligations,omitempty"`
	EngineVersion     string          `json:"engine_version,omitempty"`
}

// LoanInput matches the webapp loan shape for stress/classification.
type LoanInput struct {
	ID               string          `json:"id"`
	Principal        decimal.Decimal `json:"principal"`
	AnnualRate       decimal.Decimal `json:"annual_rate"`
	TenureMonths     int             `json:"tenure_months"`
	MoratoriumMonths int             `json:"moratorium_months"`
	Secured          *bool           `json:"secured,omitempty"` // optional; if nil, treated as unsecured
}

// ClassificationResponse is the response for POST /api/v1/classification.
type ClassificationResponse struct {
	Classification string             `json:"classification"`
	Confidence     string             `json:"confidence"`
	Metrics        StressMetrics      `json:"metrics"`
	Flags          []string           `json:"flags"`
}

// StressMetrics holds computed context metrics (v1).
type StressMetrics struct {
	EMIRatio            float64 `json:"emi_ratio"`
	FragmentationIndex  float64 `json:"fragmentation_index"`
	RigidityScore       float64 `json:"rigidity_score"`
	ShockBufferMonths   float64 `json:"shock_buffer_months"`
	UnsecuredRatio      float64 `json:"unsecured_ratio"`
}

// StressMetricsResponse is the response for POST /api/v1/stress/metrics.
type StressMetricsResponse struct {
	Metrics StressMetrics `json:"metrics"`
}

// Treatment categories (v1).
const (
	ClassificationSafeToHold      = "SAFE_TO_HOLD"
	ClassificationNeedsMonitoring = "NEEDS_MONITORING"
	ClassificationAvoidAdding     = "AVOID_ADDING"
	ClassificationActivelyReduce  = "ACTIVELY_REDUCE"
)

// Flags (explainability).
const (
	FlagHighCashflowRigidity   = "HIGH_CASHFLOW_RIGIDITY"
	FlagShortTermDebtStacking  = "SHORT_TERM_DEBT_STACKING"
	FlagLowShockBuffer          = "LOW_SHOCK_BUFFER"
	FlagUnsecuredOverexposure   = "UNSECURED_OVEREXPOSURE"
)
