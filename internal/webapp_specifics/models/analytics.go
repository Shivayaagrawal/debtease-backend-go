package models

type AnalyticsResponse struct {
	RepeatRate            float64 `json:"repeat_rate"`
	StrategySwitchRate    float64 `json:"strategy_switch_rate"`
	PdfConversionRate     float64 `json:"pdf_conversion_rate"`
	CommitmentRate        float64 `json:"commitment_rate"`
	ReturningSessionsRate float64 `json:"returning_sessions_rate"`
}
