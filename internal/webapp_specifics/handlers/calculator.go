package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func HandleBaselineCalculate(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Baseline calculate request", "path", r.URL.Path, "method", r.Method)

		// Get or create session from cookie
		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			sessionID = uuid.New()
		}

		// Parse request
		var req webapp_models.BaselineCalculateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for baseline calculate", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate request
		if len(req.Loans) == 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "At least one loan is required")
			return
		}

		// Set engine version default
		engineVersion := req.EngineVersion
		if engineVersion == "" {
			engineVersion = webapp_service.DefaultEngineVersion
		}

		// Get or create session
		sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, engineVersion)
		if err != nil {
			logger.Logger.Errorw("Failed to get or create session", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process session")
			return
		}

		// Emit calculator_started event
		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "calculator_started", nil); err != nil {
			logger.Logger.Warnw("Failed to emit calculator_started event", "error", err)
			// Don't fail the request, just log
		}

		// Calculate baseline
		result, err := webapp_service.CalculateBaseline(ctx, req.Loans, engineVersion)
		if err != nil {
			logger.Logger.Errorw("Failed to calculate baseline", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to calculate")
			return
		}

		// Emit result_viewed event
		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "result_viewed", map[string]interface{}{
			"loans":          req.Loans,
			"engine_version": engineVersion,
			"screen":         "baseline_calculator",
		}); err != nil {
			logger.Logger.Warnw("Failed to emit result_viewed event", "error", err)
			// Don't fail the request, just log
		}

		// Set session cookie
		webapp_utils.SetSessionCookie(w, sessionID)

		// Return response
		utils.RespondWithJSON(w, http.StatusOK, result)
	}
}

func HandleStrategyCalculate(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Strategy calculate request", "path", r.URL.Path, "method", r.Method)

		// Get or create session from cookie
		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			sessionID = uuid.New()
		}

		// Parse request
		var req webapp_models.StrategyCalculateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for strategy calculate", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate request
		if len(req.Loans) == 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "At least one loan is required")
			return
		}

		if req.Strategy != "AVALANCHE" && req.Strategy != "SNOWBALL" {
			utils.RespondWithError(w, http.StatusBadRequest, "Strategy must be AVALANCHE or SNOWBALL")
			return
		}

		// Set engine version default
		engineVersion := req.EngineVersion
		if engineVersion == "" {
			engineVersion = webapp_service.DefaultEngineVersion
		}

		// Get or create session
		sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, engineVersion)
		if err != nil {
			logger.Logger.Errorw("Failed to get or create session", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process session")
			return
		}

		// Emit strategy_changed event
		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "strategy_changed", map[string]interface{}{
			"strategy":       req.Strategy,
			"monthly_extra":  req.MonthlyExtra,
			"loans":          req.Loans,
			"engine_version": engineVersion,
			"screen":         "strategy_calculator",
		}); err != nil {
			logger.Logger.Warnw("Failed to emit strategy_changed event", "error", err)
			// Don't fail the request, just log
		}

		// Calculate strategy
		result, err := webapp_service.CalculateStrategy(ctx, req.Loans, req.Strategy, req.MonthlyExtra, engineVersion)
		if err != nil {
			logger.Logger.Errorw("Failed to calculate strategy", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to calculate")
			return
		}

		// Emit result_viewed event
		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "result_viewed", map[string]interface{}{
			"strategy":       req.Strategy,
			"monthly_extra":  req.MonthlyExtra,
			"loans":          req.Loans,
			"engine_version": engineVersion,
			"screen":         "strategy_calculator",
		}); err != nil {
			logger.Logger.Warnw("Failed to emit result_viewed event", "error", err)
			// Don't fail the request, just log
		}

		// Set session cookie
		webapp_utils.SetSessionCookie(w, sessionID)

		// Return response
		utils.RespondWithJSON(w, http.StatusOK, result)
	}
}
