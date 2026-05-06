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

func HandleCreateEvent(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Create event request", "path", r.URL.Path, "method", r.Method)

		// Get session_id from cookie
		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			logger.Logger.Warnw("Missing or invalid session_id in cookie")
			utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id")
			return
		}

		// Validate session exists
		exists, err := webapp_service.SessionExists(ctx, cfg.DB, sessionID)
		if err != nil {
			logger.Logger.Errorw("Failed to check session existence", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to validate session")
			return
		}
		if !exists {
			logger.Logger.Warnw("Session does not exist", "session_id", sessionID)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid session_id")
			return
		}

		// Parse request
		var req webapp_models.EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for event", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate event_type
		if req.EventType == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "event_type is required")
			return
		}

		// Emit event (service will validate event_type against allowlist)
		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, req.EventType, req.Metadata); err != nil {
			if err == webapp_service.ErrInvalidEventType {
				logger.Logger.Warnw("Invalid event type", "event_type", req.EventType)
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid event_type")
				return
			}
			logger.Logger.Errorw("Failed to emit event", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create event")
			return
		}

		// Return success
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
	}
}
