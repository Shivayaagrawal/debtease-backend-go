package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SessionResponse is the JSON response for GET /api/v1/session.
type SessionResponse struct {
	SessionID     string    `json:"session_id"`
	EngineVersion string    `json:"engine_version"`
	CreatedAt     time.Time `json:"created_at"`
}

func HandleGetSession(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Get session request", "path", r.URL.Path, "method", r.Method)

		var sessionID uuid.UUID
		if id, err := webapp_utils.GetSessionID(r); err == nil && id != uuid.Nil {
			sessionID = id
		} else if q := r.URL.Query().Get("session_id"); q != "" {
			if id, err := uuid.Parse(q); err == nil {
				sessionID = id
			}
		}
		if sessionID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id (cookie or query)")
			return
		}

		session, err := webapp_service.GetSessionByID(ctx, cfg.DB, sessionID)
		if err != nil {
			logger.Logger.Errorw("Failed to get session", "error", err, "session_id", sessionID)
			utils.RespondWithError(w, http.StatusNotFound, "Session not found")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, SessionResponse{
			SessionID:     session.SessionID.String(),
			EngineVersion: session.EngineVersion,
			CreatedAt:     session.CreatedAt,
		})
	}
}
