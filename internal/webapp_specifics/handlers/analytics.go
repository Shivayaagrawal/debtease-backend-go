package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	"net/http"
)

func HandleGetAnalytics(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Get analytics", "path", r.URL.Path, "method", r.Method)

		// Get all analytics rates
		repeatRate, err := webapp_service.GetRepeatCalculationRate(ctx, cfg.DB)
		if err != nil {
			logger.Logger.Errorw("Failed to get repeat calculation rate", "error", err)
			// Continue with zero value
		}

		strategySwitchRate, err := webapp_service.GetStrategySwitchRate(ctx, cfg.DB)
		if err != nil {
			logger.Logger.Errorw("Failed to get strategy switch rate", "error", err)
			// Continue with zero value
		}

		pdfConversionRate, err := webapp_service.GetPdfConversionRate(ctx, cfg.DB)
		if err != nil {
			logger.Logger.Errorw("Failed to get PDF conversion rate", "error", err)
			// Continue with zero value
		}

		commitmentRate, err := webapp_service.GetCommitmentRate(ctx, cfg.DB)
		if err != nil {
			logger.Logger.Errorw("Failed to get commitment rate", "error", err)
			// Continue with zero value
		}

		returningSessionsRate, err := webapp_service.GetReturningSessionsRate(ctx, cfg.DB)
		if err != nil {
			logger.Logger.Errorw("Failed to get returning sessions rate", "error", err)
			// Continue with zero value
		}

		// Return aggregated metrics
		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AnalyticsResponse{
			RepeatRate:            repeatRate,
			StrategySwitchRate:    strategySwitchRate,
			PdfConversionRate:     pdfConversionRate,
			CommitmentRate:        commitmentRate,
			ReturningSessionsRate: returningSessionsRate,
		})
	}
}

