package stress

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"
)

func HandlePostClassification(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = ctx
		logger.Logger.Infow("POST classification", "path", r.URL.Path, "method", r.Method)

		var req ClassificationRequest
		if r.Body == nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Request body required")
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		monthlyIncome := decimal.Zero
		fixedObligations := decimal.Zero
		if req.MonthlyIncome != nil {
			monthlyIncome = *req.MonthlyIncome
		}
		if req.FixedObligations != nil {
			fixedObligations = *req.FixedObligations
		}

		metrics := ComputeMetrics(req.Loans, monthlyIncome, fixedObligations)
		classification, confidence, flags := Classify(metrics)

		utils.RespondWithJSON(w, http.StatusOK, ClassificationResponse{
			Classification: classification,
			Confidence:     confidence,
			Metrics:        metrics,
			Flags:          flags,
		})
	}
}

func HandlePostStressMetrics(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Infow("POST stress metrics", "path", r.URL.Path, "method", r.Method)

		var req ClassificationRequest
		if r.Body == nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Request body required")
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		monthlyIncome := decimal.Zero
		fixedObligations := decimal.Zero
		if req.MonthlyIncome != nil {
			monthlyIncome = *req.MonthlyIncome
		}
		if req.FixedObligations != nil {
			fixedObligations = *req.FixedObligations
		}

		metrics := ComputeMetrics(req.Loans, monthlyIncome, fixedObligations)
		utils.RespondWithJSON(w, http.StatusOK, StressMetricsResponse{Metrics: metrics})
	}
}
