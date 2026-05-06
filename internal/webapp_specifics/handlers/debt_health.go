package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"DebtEase/internal/stress"
	"net/http"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DebtHealthResponse is the response for GET /api/v1/debt-health.
type DebtHealthResponse struct {
	OverallHealthScore int      `json:"overall_health_score"`
	DebtCount          int      `json:"debt_count"`
	Classification     string   `json:"classification,omitempty"`
}

func HandleGetDebtHealth(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("GET debt health", "path", r.URL.Path, "method", r.Method)

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			if q := r.URL.Query().Get("session_id"); q != "" {
				if id, parseErr := uuid.Parse(q); parseErr == nil {
					sessionID = id
				}
			}
		}
		if sessionID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id (cookie or query)")
			return
		}

		steps, err := webapp_service.GetFinanceSteps(ctx, cfg.DB, sessionID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load finance steps")
			return
		}

		resp := DebtHealthResponse{DebtCount: 0, OverallHealthScore: 100}
		if steps == nil || len(steps.FinanceSteps.Loans) == 0 {
			utils.RespondWithJSON(w, http.StatusOK, resp)
			return
		}

		monthlyIncome := decimal.Zero
		fixedObligations := decimal.Zero
		if steps.FinanceSteps.MonthlyIncome != nil {
			monthlyIncome = *steps.FinanceSteps.MonthlyIncome
		}
		if steps.FinanceSteps.FixedObligations != nil {
			fixedObligations = *steps.FinanceSteps.FixedObligations
		}

		loans := make([]stress.LoanInput, len(steps.FinanceSteps.Loans))
		for i, l := range steps.FinanceSteps.Loans {
			loans[i] = stress.LoanInput{
				ID:               l.ID,
				Principal:        l.Principal,
				AnnualRate:       l.AnnualRate,
				TenureMonths:     l.TenureMonths,
				MoratoriumMonths: l.MoratoriumMonths,
			}
		}
		metrics := stress.ComputeMetrics(loans, monthlyIncome, fixedObligations)
		classification, _, _ := stress.Classify(metrics)

		score := classificationToScore(classification)
		resp.DebtCount = len(loans)
		resp.OverallHealthScore = score
		resp.Classification = classification
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func classificationToScore(c string) int {
	switch c {
	case stress.ClassificationSafeToHold:
		return 80
	case stress.ClassificationNeedsMonitoring:
		return 60
	case stress.ClassificationAvoidAdding:
		return 40
	case stress.ClassificationActivelyReduce:
		return 20
	default:
		return 60
	}
}
