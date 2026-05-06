package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/models"
	"DebtEase/internal/utils"
	"context"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

func HandleGetDashboard(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx := context.Background()

		// Fetch dashboard summary
		summary, err := cfg.DB.GetDashboardSummary(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("GetDashboardSummary failed", "error", err, "userID", userID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch dashboard summary")
			return
		}

		// Fetch upcoming EMIs
		upcomingEMIs, err := cfg.DB.GetUpcomingEMIs(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("GetUpcomingEMIs failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch upcoming EMIs")
			return
		}

		// Fetch payment history summary
		paymentHistory, err := cfg.DB.GetPaymentHistorySummary(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("GetPaymentHistorySummary failed", "error", err, "userID", userID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch payment history")
			return
		}

		// Convert dashboard summary
		response := models.DashboardResponse{
			RepaymentProgress: models.RepaymentProgress{},
			DebtHealth: models.DebtHealth{
				TotalActiveDebts: summary.TotalActiveDebts,
				OverdueCount:     summary.OverdueCount,
				HighRiskCount:    summary.HighRiskCount,
			},
			UpcomingEMIs:   []models.UpcomingEMIResponse{},
			PaymentHistory: models.PaymentHistorySummary{},
		}

		// Convert interface{} to decimal.Decimal for monetary values
		response.TotalOutstanding = convertToDecimal(summary.TotalOutstanding)
		response.RepaymentProgress.TotalPrincipal = convertToDecimal(summary.TotalPrincipal)
		response.RepaymentProgress.TotalPaid = convertToDecimal(summary.TotalPaid)
		response.MonthlyMinimumPayment = convertToDecimal(summary.TotalMinimumPayment)

		// Convert percentages (int32) to decimal.Decimal
		// repayment_percentage is multiplied by 10000 in SQL to preserve 2 decimal places (1130 = 11.30%)
		// Divide by 100 to get the percentage value as decimal
		response.RepaymentProgress.RepaymentPercentage = decimal.NewFromInt32(summary.RepaymentPercentage).Div(decimal.NewFromInt(100))
		// overall_health_score is out of 100, so keep as-is
		response.DebtHealth.OverallHealthScore = decimal.NewFromInt32(summary.OverallHealthScore)

		// Convert interest rate: DECIMAL(6,5) with 5 decimal places
		// sqlc converts the aggregated result to int32, likely scaled by 100000
		// Example: 0.15500 (15.5%) stored in DB -> 15500 in int32 -> divide by 100000 to get 0.155
		response.AverageInterestRate = decimal.NewFromInt32(summary.WeightedAvgInterestRate).Div(decimal.NewFromInt(100000))

		// Handle months to debt free
		if summary.MonthsToDebtFree != nil {
			months := convertToInt32(summary.MonthsToDebtFree)
			if months != nil {
				response.MonthsToDebtFree = months
				// Calculate expected debt free date
				expectedDate := time.Now().AddDate(0, int(*months), 0)
				response.ExpectedDebtFree = &expectedDate
			}
		}

		// Convert upcoming EMIs
		for _, emi := range upcomingEMIs {
			emiResponse := models.UpcomingEMIResponse{
				ID:                 emi.ID.String(),
				Name:               emi.Name,
				Type:               string(emi.Type),
				OutstandingBalance: emi.OutstandingBalance,
				AccruedInterest:    emi.AccruedInterest,
				PaymentAmount:      emi.PaymentAmount,
				DueDate:            emi.DueDate.Format("2006-01-02"),
				IsOverdue:          emi.IsOverdue.Bool && emi.IsOverdue.Valid,
			}

			if emi.Lender.Valid {
				emiResponse.Lender = &emi.Lender.String
			}
			if emi.PaymentDueDay.Valid {
				emiResponse.PaymentDueDay = &emi.PaymentDueDay.Int32
			}
			if emi.LateFees.Valid && emi.LateFees.String != "" {
				lateFees, err := decimal.NewFromString(emi.LateFees.String)
				if err == nil {
					emiResponse.LateFees = &lateFees
				}
			}

			// Convert days_until_due from interface{} to int32
			if emi.DaysUntilDue != nil {
				if days, ok := emi.DaysUntilDue.(int64); ok {
					emiResponse.DaysUntilDue = int32(days)
				} else if days, ok := emi.DaysUntilDue.(int32); ok {
					emiResponse.DaysUntilDue = days
				}
			}

			response.UpcomingEMIs = append(response.UpcomingEMIs, emiResponse)
		}

		// Convert payment history summary
		response.PaymentHistory.PaidLast30Days = convertToDecimal(paymentHistory.PaidLast30Days)
		response.PaymentHistory.PaidLast90Days = convertToDecimal(paymentHistory.PaidLast90Days)
		response.PaymentHistory.PaidLastYear = convertToDecimal(paymentHistory.PaidLastYear)
		response.PaymentHistory.PaymentsLast30Days = paymentHistory.PaymentsLast30Days

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// Helper function to convert interface{} to decimal.Decimal
func convertToDecimal(val interface{}) decimal.Decimal {
	if val == nil {
		return decimal.Zero
	}

	switch v := val.(type) {
	case string:
		d, err := decimal.NewFromString(v)
		if err != nil {
			logger.Logger.Warnw("Failed to parse decimal from string", "value", v, "error", err)
			return decimal.Zero
		}
		return d
	case []byte:
		d, err := decimal.NewFromString(string(v))
		if err != nil {
			logger.Logger.Warnw("Failed to parse decimal from bytes", "value", string(v), "error", err)
			return decimal.Zero
		}
		return d
	case int64:
		return decimal.NewFromInt(v)
	case int32:
		return decimal.NewFromInt32(v)
	case float64:
		return decimal.NewFromFloat(v)
	default:
		// Try to convert to string first
		if str, ok := val.(string); ok {
			d, err := decimal.NewFromString(str)
			if err != nil {
				logger.Logger.Warnw("Failed to parse decimal from interface", "value", val, "error", err)
				return decimal.Zero
			}
			return d
		}
		logger.Logger.Warnw("Unknown type for decimal conversion", "type", val, "value", val)
		return decimal.Zero
	}
}

// Helper function to convert interface{} to int32 pointer
func convertToInt32(val interface{}) *int32 {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case int32:
		return &v
	case int64:
		result := int32(v)
		return &result
	case int:
		result := int32(v)
		return &result
	default:
		logger.Logger.Warnw("Unknown type for int32 conversion", "type", val, "value", val)
		return nil
	}
}
